// ts-node -O '{"module": "commonjs"}' getmockdata.ts

import http from 'http'
import fs from 'fs'

// see: https://www.tomas-dvorak.cz/posts/nodejs-request-without-dependencies/,
// even more simplified.
const httpget = (url: string) => {
    return new Promise((resolve, reject) => {
        const request = http.get(url, (response) => {
            // handle http errors
            if (response.statusCode < 200 || response.statusCode > 299) {
                reject(new Error('status code: ' + response.statusCode));
            }
            const body = [];
            response.on('data', (chunk) => body.push(chunk));
            response.on('end', () => resolve(body.join('')));
        });
        request.on('error', (err) => reject(err))
    })
}

;console.log('fetching discovery data from local Ghostwire service at :5000...')

;(async () => {
    let data
    await httpget('http://localhost:5000/json')
        .then(result => {data = JSON.parse(result as string)})
        .catch(fail => {
            console.log('error calling discovery REST API:', fail)
            process.exit(1)
        })
    data.metadata.hostname = 'mockbox'
    const initialnetnsid = data['fake-roots'][0].netnsid
    const initialnetns = data['network-namespaces'].find(netns => netns.netnsid === initialnetnsid)
    Object.keys(initialnetns['transport-ports']).forEach(family =>
            initialnetns['transport-ports'][family] = initialnetns['transport-ports'][family].filter(port => {
                const localport: number = port['local-port']
                return !!([22, 53, 68, 631, 5000, 5010].find(p => p === localport))
            }));
    const initdcntr = initialnetns.containers.find(cntr => cntr.pid === 1)
    const hn = initdcntr.dns['uts-hostname']
    const sl = initdcntr.dns.searchlist[0]
    data['network-namespaces'].forEach(netns =>
        netns.containers.forEach(cntr => {
            cntr.dns.searchlist = !cntr.dns.searchlist.find(sle => sle === sl) ? cntr.dns.searchlist : "mock.example.org"
            cntr.dns['etc-hosts'] = cntr.dns['etc-hosts'].map(nameaddr => ({
                name: nameaddr.name !== hn ? nameaddr.name : 'mockbox',
                address: nameaddr.address,
            }))
            cntr.dns['uts-hostname'] = cntr.dns['uts-hostname'] !== hn ? cntr.dns['uts-hostname'] : ["mockbox"]
            cntr.dns['etc-hostname'] = cntr.dns['etc-hostname'] !== hn ? cntr.dns['etc-hostname'] : ["mockbox"]
        }))
    
    const outname = process.argv[2] || 'mockdata.json'
    console.log(`writing ${outname}...`)
    fs.writeFileSync(outname, JSON.stringify(data, null, 4))
})();

console.log('Done.')
