
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/private';
 * 
 * console.log(ENVIRONMENT); // => "production"
 * console.log(PUBLIC_BASE_URL); // => throws error during build
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/private' {
	export const XDG_DATA_DIRS: string;
	export const LESS_TERMCAP_se: string;
	export const ORCA_AGENT_HOOK_VERSION: string;
	export const NVM_RC_VERSION: string;
	export const LESS_TERMCAP_ue: string;
	export const XDG_SESSION_TYPE: string;
	export const GIT_CONFIG_COUNT: string;
	export const NVM_CD_FLAGS: string;
	export const LESS_TERMCAP_md: string;
	export const GDK_BACKEND: string;
	export const SSH_AUTH_SOCK: string;
	export const XDG_SESSION_DESKTOP: string;
	export const BUN_INSTALL: string;
	export const FORCE_HYPERLINK: string;
	export const SHELL: string;
	export const QT_QPA_PLATFORMTHEME: string;
	export const AGENT: string;
	export const GIT_CONFIG_KEY_0: string;
	export const XDG_MENU_PREFIX: string;
	export const XDG_CURRENT_DESKTOP: string;
	export const FC_FONTATIONS: string;
	export const COMMAND_NOT_FOUND_INSTALL_PROMPT: string;
	export const ORCA_AGENT_HOOK_PORT: string;
	export const PANEL_GDK_CORE_DEVICE_EVENTS: string;
	export const LESS_TERMCAP_me: string;
	export const XDG_CACHE_HOME: string;
	export const ORCA_TAB_ID: string;
	export const ORCA_WORKTREE_ID: string;
	export const XDG_SESSION_ID: string;
	export const GIT_TERMINAL_PROMPT: string;
	export const POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD: string;
	export const ORCA_AGENT_HOOK_ENDPOINT: string;
	export const WINDOWID: string;
	export const NMAP_PRIVILEGED: string;
	export const ORCA_PANE_KEY: string;
	export const GIT_ASKPASS: string;
	export const POWERSHELL_UPDATECHECK: string;
	export const XDG_SEAT: string;
	export const GDMSESSION: string;
	export const ORCA_PI_SOURCE_AGENT_DIR: string;
	export const LANG: string;
	export const SESSION_MANAGER: string;
	export const MODAL_PROXY_TOKEN: string;
	export const XDG_CONFIG_HOME: string;
	export const GSETTINGS_SCHEMA_DIR: string;
	export const ORCA_USER_DATA_PATH: string;
	export const QT_AUTO_SCREEN_SCALE_FACTOR: string;
	export const HOME: string;
	export const LANGUAGE: string;
	export const DOTNET_CLI_TELEMETRY_OPTOUT: string;
	export const ORCA_WORKSPACE_ID: string;
	export const USER: string;
	export const XDG_SESSION_PATH: string;
	export const ORCA_OPENCODE_CONFIG_DIR: string;
	export const ORCA_AGENT_LAUNCH_TOKEN: string;
	export const DESKTOP_SESSION: string;
	export const XDG_GREETER_DATA_DIR: string;
	export const ORCA_TERMINAL_HANDLE: string;
	export const NVM_DIR: string;
	export const COLORTERM: string;
	export const POWERSHELL_TELEMETRY_OPTOUT: string;
	export const PWD: string;
	export const OPENCODE: string;
	export const SSH_AGENT_PID: string;
	export const XDG_RUNTIME_DIR: string;
	export const GIT_CONFIG_KEY_1: string;
	export const GIT_CONFIG_VALUE_1: string;
	export const ZDOTDIR: string;
	export const OLDPWD: string;
	export const _: string;
	export const ORCA_ORIG_ZDOTDIR: string;
	export const XDG_SESSION_CLASS: string;
	export const ORCA_SHELL_READY_MARKER: string;
	export const TERM: string;
	export const XDG_VTNR: string;
	export const LOGNAME: string;
	export const GIT_CONFIG_VALUE_0: string;
	export const TERM_PROGRAM_VERSION: string;
	export const ORCA_AGENT_HOOK_TOKEN: string;
	export const LESS_TERMCAP_us: string;
	export const PATH: string;
	export const LESS_TERMCAP_so: string;
	export const COLORFGBG: string;
	export const CHROME_DESKTOP: string;
	export const TERM_PROGRAM: string;
	export const XAUTHORITY: string;
	export const DEEPSEEK_API_KEY: string;
	export const OPENCODE_PID: string;
	export const SSH_ASKPASS: string;
	export const XDG_CONFIG_DIRS: string;
	export const LS_COLORS: string;
	export const DISPLAY: string;
	export const ORCA_APP_VERSION: string;
	export const ORCA_AGENT_HOOK_ENV: string;
	export const XDG_SEAT_PATH: string;
	export const GCM_INTERACTIVE: string;
	export const OPENCODE_CONFIG_DIR: string;
	export const SHLVL: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const LESS_TERMCAP_mb: string;
	export const QT_ACCESSIBILITY: string;
	export const NODE_ENV: string;
}

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/public';
 * 
 * console.log(ENVIRONMENT); // => throws error during build
 * console.log(PUBLIC_BASE_URL); // => "http://site.com"
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * 
 * console.log(env.ENVIRONMENT); // => "production"
 * console.log(env.PUBLIC_BASE_URL); // => undefined
 * ```
 */
declare module '$env/dynamic/private' {
	export const env: {
		XDG_DATA_DIRS: string;
		LESS_TERMCAP_se: string;
		ORCA_AGENT_HOOK_VERSION: string;
		NVM_RC_VERSION: string;
		LESS_TERMCAP_ue: string;
		XDG_SESSION_TYPE: string;
		GIT_CONFIG_COUNT: string;
		NVM_CD_FLAGS: string;
		LESS_TERMCAP_md: string;
		GDK_BACKEND: string;
		SSH_AUTH_SOCK: string;
		XDG_SESSION_DESKTOP: string;
		BUN_INSTALL: string;
		FORCE_HYPERLINK: string;
		SHELL: string;
		QT_QPA_PLATFORMTHEME: string;
		AGENT: string;
		GIT_CONFIG_KEY_0: string;
		XDG_MENU_PREFIX: string;
		XDG_CURRENT_DESKTOP: string;
		FC_FONTATIONS: string;
		COMMAND_NOT_FOUND_INSTALL_PROMPT: string;
		ORCA_AGENT_HOOK_PORT: string;
		PANEL_GDK_CORE_DEVICE_EVENTS: string;
		LESS_TERMCAP_me: string;
		XDG_CACHE_HOME: string;
		ORCA_TAB_ID: string;
		ORCA_WORKTREE_ID: string;
		XDG_SESSION_ID: string;
		GIT_TERMINAL_PROMPT: string;
		POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD: string;
		ORCA_AGENT_HOOK_ENDPOINT: string;
		WINDOWID: string;
		NMAP_PRIVILEGED: string;
		ORCA_PANE_KEY: string;
		GIT_ASKPASS: string;
		POWERSHELL_UPDATECHECK: string;
		XDG_SEAT: string;
		GDMSESSION: string;
		ORCA_PI_SOURCE_AGENT_DIR: string;
		LANG: string;
		SESSION_MANAGER: string;
		MODAL_PROXY_TOKEN: string;
		XDG_CONFIG_HOME: string;
		GSETTINGS_SCHEMA_DIR: string;
		ORCA_USER_DATA_PATH: string;
		QT_AUTO_SCREEN_SCALE_FACTOR: string;
		HOME: string;
		LANGUAGE: string;
		DOTNET_CLI_TELEMETRY_OPTOUT: string;
		ORCA_WORKSPACE_ID: string;
		USER: string;
		XDG_SESSION_PATH: string;
		ORCA_OPENCODE_CONFIG_DIR: string;
		ORCA_AGENT_LAUNCH_TOKEN: string;
		DESKTOP_SESSION: string;
		XDG_GREETER_DATA_DIR: string;
		ORCA_TERMINAL_HANDLE: string;
		NVM_DIR: string;
		COLORTERM: string;
		POWERSHELL_TELEMETRY_OPTOUT: string;
		PWD: string;
		OPENCODE: string;
		SSH_AGENT_PID: string;
		XDG_RUNTIME_DIR: string;
		GIT_CONFIG_KEY_1: string;
		GIT_CONFIG_VALUE_1: string;
		ZDOTDIR: string;
		OLDPWD: string;
		_: string;
		ORCA_ORIG_ZDOTDIR: string;
		XDG_SESSION_CLASS: string;
		ORCA_SHELL_READY_MARKER: string;
		TERM: string;
		XDG_VTNR: string;
		LOGNAME: string;
		GIT_CONFIG_VALUE_0: string;
		TERM_PROGRAM_VERSION: string;
		ORCA_AGENT_HOOK_TOKEN: string;
		LESS_TERMCAP_us: string;
		PATH: string;
		LESS_TERMCAP_so: string;
		COLORFGBG: string;
		CHROME_DESKTOP: string;
		TERM_PROGRAM: string;
		XAUTHORITY: string;
		DEEPSEEK_API_KEY: string;
		OPENCODE_PID: string;
		SSH_ASKPASS: string;
		XDG_CONFIG_DIRS: string;
		LS_COLORS: string;
		DISPLAY: string;
		ORCA_APP_VERSION: string;
		ORCA_AGENT_HOOK_ENV: string;
		XDG_SEAT_PATH: string;
		GCM_INTERACTIVE: string;
		OPENCODE_CONFIG_DIR: string;
		SHLVL: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		LESS_TERMCAP_mb: string;
		QT_ACCESSIBILITY: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://example.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.ENVIRONMENT); // => undefined, not public
 * console.log(env.PUBLIC_BASE_URL); // => "http://example.com"
 * ```
 * 
 * ```
 * 
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
