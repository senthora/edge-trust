package config

import "os"

type EnvVars struct {
	APIURL              string
	ProxySourcesPath    string
	OriginAllowlistPath string
	ReloadSignalPath    string
	StateJSONPath       string
	HCSignalPath        string
}

func LoadEnv() EnvVars {
	return EnvVars{
		APIURL:              getRequiredEnv("CF_API_URL"),
		ProxySourcesPath:    getRequiredEnv("NGINX_PROXY_SOURCES_PATH"),
		OriginAllowlistPath: getRequiredEnv("NGINX_ORIGIN_ALLOWLIST_PATH"),
		ReloadSignalPath:    getRequiredEnv("NGINX_RELOAD_SIGNAL_PATH"),
		StateJSONPath:       getOptionalEnv("STATE_JSON_PATH", "/var/lib/edge-trust/state.json"),
		HCSignalPath:        getOptionalEnv("HEALTH_SIGNAL_PATH", "/var/run/edge-trust/.alive"),
	}
}

func getRequiredEnv(name string) string {
	env, exists := os.LookupEnv(name)
	if !exists {
		panic(name + " environment variable not set")
	}
	return env
}

func getOptionalEnv(name string, defaultValue string) string {
	env, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}
	return env
}
