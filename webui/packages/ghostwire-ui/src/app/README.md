# The `app/` Folder

This `app/` folder finally defines the Ghostwire App (in form of the `App`
component), by stitching together the main elements of the application. Namely,
the ...

- application bar with app title, drawer menu button, and a few action buttons,
- swipeable drawer for navigation,
- routing for selecting the correct view components based on the current route. For details on the views, please refer to the `views/` folder.

Additionally, the App also provides the infrastructure in form of...

- snack bar for showing (error) toasts, such as when the discovery service
  fails.
- `Discovery` component for fetching and providing the latest and greates
  Ghostwire discovery results.
- [jotai](https://github.com/pmndrs/jotai) state provider for distributing
  application state to all the different places in the application.

That's it.

## Themes and Branding

`createGhostwireApp(config)` returns the App component configured with
different MUI theme options for light and dark mode, as well as a different
branding (name, icon, favicon, about content, legal links, and version). All
configuration elements default to Ghostwire's defaults; please see
`config.ts`, `themes.ts`, and `components/branding/` for details.

The OSS app in `packages/app/` creates the app using the configuration from
the `ghostwire-app-config` module, which defaults to its `src/defaultconfig.ts`
unless the `VITE_APP_CONFIG` environment variable points to a different module
at build time.
