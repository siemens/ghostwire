# server.py is used in development of the Ghostwire UI mark II to test and
# emulate serving a static optimized production build -- but without the need to
# rebuild the Ghostwire container for this.
#
# 1. run the Ghostwire discovery service ("make start" inside the root directory
#    of the Ghostwire repository).
# 2. run "python3 -m fakegw.server" inside the root directory of this
#    repository.
# 3. build a static optimized build by running "yarn build" inside the root
#    directory of this repository.
# 4. navigate your browser to port :5555
# 5. rinse and repeat from step 3.
#
# Structure:
#
# browser --> :5555 -+--> :5000/discovery
#                    `--> build/ static production build files

from flask import Flask, request, Response, send_from_directory, abort
import requests
from urllib.parse import urlparse
import re
import json
from os.path import abspath, exists, join
import logging

# Switch on logging to stdout from the debug level on.
logging.basicConfig(level=logging.DEBUG)

BUILD_DIR = 'build'
ORIGINAL_URL_HEADER = 'X-Forwarded-Uri'
ENABLE_CAPTURE_HEADER = 'Enable-Monolith'


# Switch off any flask-internal in-convenience handling of "/static/..." URL
# paths, so we get full control over any servings.
app = Flask('gwui',
            static_url_path=None,
            root_path=None,
            static_folder=None)

build_dir = abspath(BUILD_DIR)
# Argh! Don't be greedy, use "*?" instead of "*"...
basename_re = re.compile(r'(<base href=").*?("/>)')
dynvars_re = re.compile(r'(<script>window\.dynvars=){}(</script>)')


@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
# If the request is for an existing asset, then serve it; otherwise, always
# serve our index.html.
def catch_all(path):
    # Does the path we are finally seeing here must correctly address an asset
    # to be served. Or else we serve the interpolated index.html.
    if path != '' and exists(join(build_dir, path)):
        return send_from_directory(build_dir, path)

    # Did an upstream (ingress) reverse proxy tell us what the original URI
    # actually was? If not, take what we've finally seen here.
    uri = request.headers.get(ORIGINAL_URL_HEADER)
    if not uri:
        uri = request.url
    # Now find the base path (~prefix) so we can correctly interpolate the base
    # href when serving the SPA. We only care about the path aspect.
    uri_path = urlparse(uri).path
    base = uri_path[:-len(path)-1] if uri_path.endswith('/'+path) else ''
    # Ensure that base path always ends with '/' -- or try to find out that the
    # base href will otherwise being treated to a "dirname" operation. Oh, well.
    if not base.endswith('/'):
        base += '/'
    # Sanitize base so it cannot play havoc with our re.sub replacement.
    base = base.replace('\\', '\\\\')

    logging.info(f'{ORIGINAL_URL_HEADER}: {uri}')

    # do we get signalled to enable capture links?
    capture_links = request.headers.get(ENABLE_CAPTURE_HEADER) is not None

    # Serve index.html with the dynamically rewritten (interpolated) base href.
    # This is desperately needed by the client-side SPA in order to correctly
    # address its assets and the JSON REST API, even in the face of HTML5 DOM
    # routing and rewriting proxies.
    logging.info(f'serving dynamic index.html with base "{base}" on "/{path}"')
    dynvars = json.dumps({
        'enableMonolith': capture_links
    })
    with open(join(build_dir, 'index.html')) as f:
        indexhtml = f.read()
    return Response(
        dynvars_re.sub(
            f'\\1{dynvars}\\2',
            basename_re.sub(f'\\1{base}\\2', indexhtml)),
        mimetype='text/html')


@app.route('/json')
# Forward discovery requests to the real discovery service on this machine.
def discover():
    r = requests.get('http://localhost:5000/json')
    return Response(r.content, mimetype='application/json')


if __name__ == '__main__':
    app.run(host='::', port=5555, debug=True)
