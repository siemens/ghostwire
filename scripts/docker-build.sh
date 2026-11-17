#!/bin/bash

dockerfile="$1"
args=( "${@:2}" )

# Builds a container image using the specified Dockerfile, with the current
# directory being the Ghostwire module and main build context. In Go workspace
# mode, all other workspace modules get passed as additional build contexts
# "bctxN", together with their in-workspace directories "MODN" and the list of
# all workspace module directories "WSDISKPATHS". Inside the build, the Ghostwire
# module lives in "./ghostwire" and each other module in "./<directory name>".
#
# Use the GOWORK environment variable to explicitly specify the go.work file to
# use. Additionally, set the MAINPKG environment variable to the (host) path of
# the main package to build, if not "./cmd/gostwire"; this gets translated into
# its path inside the build.

MAINUSE=./ghostwire # in-build directory of the Ghostwire module
EMPTYCONTEXT=.emptyctx
NUMCONTEXTS=9

# find out if we are in workspace mode -- and it we are, then the list of
# modules actually used.
mkdir -p ${EMPTYCONTEXT}
trap 'rm -rf -- "${EMPTYCONTEXT}"' EXIT

contexts=()
workspace_details=$(go work edit --json)
if [[ ${workspace_details} ]]; then
    goworkdir=$(dirname $(go env GOWORK))
    echo "found workspace" ${goworkdir}
    diskpaths=$(echo ${workspace_details} | jq --raw-output '.Use | .[]? | .DiskPath')
    echo "modules used in workspace:" ${diskpaths}
    wsdiskpaths=( "${MAINUSE}" )
    while IFS= read -r module; do
        modpath="${module}"
        [[ "${modpath}" == /* ]] || modpath="${goworkdir}/${modpath}"
        if [[ "$(realpath ${modpath})" == "$(realpath .)" ]]; then
            echo "  🏠" ${module};
        else
            relcontext=$(realpath --relative-to="." ${modpath})
            contexts+=( ${relcontext} )
            wsdiskpaths+=( "./$(basename $(realpath ${relcontext}))" )
            echo "  🧩" ${module} "» 📁" ${relcontext}
        fi
    done <<< ${diskpaths}
    diskpaths="${wsdiskpaths[*]}"
else
    diskpaths="${MAINUSE}"
fi

buildctxargs=()
buildargs=()
ctxno=1
for ctx in "${contexts[@]}"; do
    buildctxargs+=( "--build-context=bctx${ctxno}=${ctx}" )
    buildargs+=( "--build-arg=MOD${ctxno}=./$(basename $(realpath ${ctx}))/" )
    ((ctxno=ctxno+1))
done
for ((;ctxno<=NUMCONTEXTS;ctxno++)); do
    buildctxargs+=( "--build-context=bctx${ctxno}=${EMPTYCONTEXT}" )
done

# Translate the main package path given in MAINPKG, as seen from the current
# directory on the host, into the main package path inside the build, relative
# to the Ghostwire module directory there (/ws/ghostwire). Inside the build,
# each other workspace module lives next to the Ghostwire module in a directory
# named after its host directory, regardless of where it lives on the host. For
# instance, "./examples/downstream/cmd/gostwire" inside the workspace module
# "./examples/downstream" becomes "../downstream/cmd/gostwire". The main package
# must be inside one of the workspace modules, otherwise the build is aborted
# early instead of failing late inside the build.
if [[ -n "${MAINPKG}" ]]; then
    # Check the other modules first, as they might be nested inside the
    # Ghostwire module directory.
    mainpkg=$(realpath "${MAINPKG}")
    inbuild=""
    for ctx in "${contexts[@]}"; do
        ctxpath=$(realpath ${ctx})
        if [[ "${mainpkg}/" == "${ctxpath}/"* ]]; then
            inbuild="../$(basename ${ctxpath})/$(realpath --relative-to=${ctxpath} ${mainpkg})"
            break
        fi
    done
    if [[ -z "${inbuild}" && "${mainpkg}/" == "$(realpath .)/"* ]]; then
        inbuild="./$(realpath --relative-to=. ${mainpkg})"
    fi
    if [[ -z "${inbuild}" ]]; then
        echo "main package ${MAINPKG} is not inside any (workspace) module" >&2
        exit 1
    fi
    echo "main package:" ${MAINPKG} "» 📦" ${inbuild}
    buildargs+=( "--build-arg=MAINPKG=${inbuild}" )
fi
echo "args:" ${buildctxargs[*]} ${buildargs[*]}
echo "build inside:" ${CWD}

docker buildx build \
    -f "${dockerfile}" \
    "${buildargs[@]}" \
    "${buildctxargs[@]}" \
    --build-arg=WSDISKPATHS="$(echo ${diskpaths})" \
    "${args[@]}" \
    .
