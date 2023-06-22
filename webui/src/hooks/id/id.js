// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useState } from 'react';

// We simply create a series of id numbers, whenenver one is requested. Using
// the state hook we then ensure that a functional component gets its own stable
// id, yet multiple components get their own stable individual ids. Why 42? Oh,
// read the pop classics!
let someId = 42;

// Returns a unique id using the given prefix or default of 'id-'.
const useId = (prefix = 'id-') => {
    // Only calculate a new id if useState really needs an initial state;
    // otherwise, skip it to keep things stable over the livetime of a single
    // component.
    const [ id ] = useState(() => prefix + (someId++).toString());
    return id;
};

export default useId;
