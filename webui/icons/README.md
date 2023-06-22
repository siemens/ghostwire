# Icon SVG Source Files

This directory holds the original SVG source files for the icons used in the
Ghostwire UI. As these SVG files contain a good deal of meta information, not
least from the graphic editor Inkscape, they are not suitable for direct
consumption in the UI. Moreover, we prefer to have Typescript components instead
of "plain" SVG.

To convert and update the icons, simply run in the repository top-level directory:

```bash
(cd webui && yarn icons)
```

For the conversion to work as expected, you should ensure to combine all of the
icon into a *single* path (Ctrl-+ will be your friend).

## License and Copyright

(c) Siemens AG 2023

[SPDX-License-Identifier: MIT](../../LICENSE)
