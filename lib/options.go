package cloudscraper

import (
	"github.com/Advik-B/cloudscraper/lib/captcha"
	"github.com/Advik-B/cloudscraper/lib/proxy"
	"github.com/Advik-B/cloudscraper/lib/stealth"
	"github.com/Advik-B/cloudscraper/lib/user_agent"
	"time"
)

// Options holds all configuration for the scraper.
type Options struct {
	MaxRetries             int
	Delay                  time.Duration
	AutoRefreshOn403       bool
	SessionRefreshInterval time.Duration
	Max403Retries          int
	Browser                useragent.Config
	RotateTlsCiphers       bool
	CaptchaSolver          captcha.Solver
	Proxies                []string
	ProxyOptions           struct {
		Strategy proxy.Strategy
		BanTime  time.Duration
	}
	Stealth stealth.Options
}

// ScraperOption configures a Scraper.
type ScraperOption func(*Options)

// WithBrowser configures the browser profile to use.
func WithBrowser(cfg useragent.Config) ScraperOption {
	return func(o *Options) {
		o.Browser = cfg
	}
}

// WithCaptchaSolver configures a captcha solver.
func WithCaptchaSolver(solver captcha.Solver) ScraperOption {
	return func(o *Options) {
		o.CaptchaSolver = solver
	}
}

// WithProxies configures the proxy manager.
func WithProxies(proxyURLs []string, strategy proxy.Strategy, banTime time.Duration) ScraperOption {
	return func(o *Options) {
		o.Proxies = proxyURLs
		o.ProxyOptions.Strategy = strategy
		o.ProxyOptions.BanTime = banTime
	}
}

// WithStealth configures the stealth mode options.
func WithStealth(opts stealth.Options) ScraperOption {
	return func(o *Options) {
		o.Stealth = opts
	}
}

// WithSessionConfig configures session handling.
func WithSessionConfig(refreshOn403 bool, interval time.Duration, maxRetries int) ScraperOption {
	return func(o *Options) {
		o.AutoRefreshOn403 = refreshOn403
		o.SessionRefreshInterval = interval
		o.Max403Retries = maxRetries
	}
}

// WithDelay sets a fixed delay between requests (used by StealthMode if HumanLikeDelays is false).
func WithDelay(d time.Duration) ScraperOption {
	return func(o *Options) {
		o.Delay = d
	}
}
