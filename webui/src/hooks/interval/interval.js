/*
Code snippets included in posts on the site are licensed under more permissive terms than general post content.
Please see LICENSE-posts for the terms regarding general post content.

Code snippets may be used under the terms of the MIT License:

MIT License

Copyright (c) 2020 Dan Abramov and the contributors.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

Modifications by Siemens
*/
import { useEffect, useRef } from 'react'

// useInterval is the declarative cousin to JS' setInternval, using react hooks.
// See: https://overreacted.io/making-setinterval-declarative-with-react-hooks/;
// the posted code is licensed under the MIT license, see:
// https://github.com/gaearon/overreacted.io/blob/master/LICENSE-code-snippets.
const useInterval = (callback, delay) => {
    const savedCallback = useRef() // no useState() here.

    // Whenever the callback or the delay changes, make sure that we remember
    // and use the most recent callback; so, no useState(), but a reference
    // instead.
    useEffect(() => {
        savedCallback.current = callback
    });

    function tick() {
        savedCallback.current()
    }

    // Whenever the delay changes, we need to set up the interval timer anew.
    // But since the callback might be changed (independently) too, we must use
    // a reference to the callback instead of the "frozen" component state.
    useEffect(() => {
        // Only set the interval timer, if the delay value isn't null; any
        // previous timer is automatically removed because we previously told
        // react how to clean it up when changing the delay value. It's like
        // parallel universes...
        if (delay !== null) {
            let id = setInterval(tick, delay)
            // ...and tell react (how) to clean up the old interval timer when
            // the delay value changes.
            return () => {
                clearInterval(id)
            }
        }
    }, [delay])
};

export default useInterval
