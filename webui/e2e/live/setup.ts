import { execFileSync } from 'node:child_process'
import path from 'node:path'

// Must match the containers in docker-compose.yaml.
export const brand = 'Ghostwire E2E'
export const containees = ['gw-e2e-alpha', 'gw-e2e-beta']
export const forwardedPort = Number(process.env.GW_E2E_FORWARDED_PORT ?? 5098)

const port = Number(process.env.GW_E2E_PORT ?? 5099)
export const baseURL = `http://127.0.0.1:${port}`

const repoRoot = path.resolve(import.meta.dirname, '../../..')
const composeFile = path.join(import.meta.dirname, 'docker-compose.yaml')
// Set GW_E2E_IMAGE to test an already built image instead of building one.
const image = process.env.GW_E2E_IMAGE ?? 'gostwire:e2e'

const run = (cmd: string, args: string[]) => execFileSync(cmd, args, {
    cwd: repoRoot,
    stdio: 'inherit',
    env: {
        ...process.env,
        GW_E2E_IMAGE: image,
        GW_E2E_PORT: String(port),
        GW_E2E_FORWARDED_PORT: String(forwardedPort),
        GW_E2E_BRAND: brand,
    },
})

const compose = (...args: string[]) =>
    run('docker', ['compose', '-p', 'ghostwire-e2e', '-f', composeFile, ...args])

const waitForDiscovery = async (timeout: number) => {
    const deadline = Date.now() + timeout
    for (; ;) {
        try {
            if ((await fetch(`${baseURL}/json`)).ok) {
                return
            }
        } catch { /* not yet listening */ }
        if (Date.now() > deadline) {
            throw new Error(`Ghostwire service at ${baseURL} not ready in time`)
        }
        await new Promise(resolve => setTimeout(resolve, 500))
    }
}

/**
 * Builds the Ghostwire service image (unless GW_E2E_IMAGE is set) and then
 * starts the service together with some well-known containers; returns the
 * global teardown removing them again.
 */
export default async function globalSetup() {
    if (!process.env.GW_E2E_IMAGE) {
        const version = execFileSync('git', ['describe', '--tags', '--always', '--dirty'],
            { cwd: repoRoot }).toString().trim()
        run('scripts/docker-build.sh', [
            'deployments/gostwire/Dockerfile',
            '--load', // into the local image store for compose, even with non-default buildx builders
            '-t', image,
            `--build-arg=VERSION=${version}`,
            '--build-context=webappsrc=./webui',
        ])
    }
    compose('up', '--detach', '--wait')
    const teardown = () => { compose('down', '--timeout', '1') }
    try {
        await waitForDiscovery(60_000)
    } catch (err) {
        teardown()
        throw err
    }
    return teardown
}
