# ingress.py is used in development to simulate the effect of an path-rewriting
# ingress reverse proxy.
#
# IMPORTANT: sets the X-Forwarded-Uri header field in forwarded request!
#
# Structure:
#
# browser --> :5556 --> :5555 --> ...

from flask import Flask, request, Response, abort
import requests
import logging


# URL (base) path where to proxy to the local GW fake server.
EDGESHARK_PATH = '/edgeshark'

ORIGINAL_URL_HEADER = 'X-Forwarded-Uri'
ENABLE_CAPTURE_HEADER = 'Enable-Monolith'

# Switch on logging to stdout from the debug level on.
logging.basicConfig(level=logging.DEBUG)

# Switch off any flask-internal in-convenience handling of "/static/..." URL
# paths, so we get full control over any servings.
app = Flask('ingress',
            static_url_path=None,
            root_path=None,
            static_folder=None)

DROP_HEADERS = [
    'content-encoding',
    'content-length',
    'transfer-encoding',
    'connection'
]


@app.route('/<path:path>', methods=['GET', 'HEAD'])
# Only forward requests for resources below the specified path; abort any
# request access elsewhere.
def forward(path):
    path = '/' + path
    logging.info(f'path: {path}')
    if not path.startswith(f'{EDGESHARK_PATH}/'):
        abort(404, 'ingress: nothing to see here')
    path = path[len(EDGESHARK_PATH):]
    logging.info(f'{ORIGINAL_URL_HEADER}: {request.url}')
    r = requests.get(
        f'http://localhost:5555{path}',
        headers={
            ORIGINAL_URL_HEADER: request.url,
            ENABLE_CAPTURE_HEADER: '2001',
        })
    headers = [(name, value) for (name, value) in r.raw.headers.items()
               if name.lower() not in DROP_HEADERS]
    return Response(r.content, r.status_code, headers)


@app.route('/')
def rude():
    abort(404, 'ingress: nothing to see here')


if __name__ == '__main__':
    app.run(host='::', port=5556, debug=True)
