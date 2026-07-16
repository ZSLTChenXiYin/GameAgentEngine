package workercli

import "github.com/spf13/pflag"

type Runner struct {
	app *app
}

func New(options Options) *Runner {
	return &Runner{app: newApp(options)}
}

func (r *Runner) BindCommonFlags(flags *pflag.FlagSet) {
	r.app.bindCommonFlags(flags)
}

func (r *Runner) BindPlayFlags(flags *pflag.FlagSet) {
	r.app.bindPlayFlags(flags)
}

func (r *Runner) BindTestFlags(flags *pflag.FlagSet) {
	r.app.bindTestFlags(flags)
}

func (r *Runner) RunServe(withPush bool, withPull bool) error {
	return r.app.runServe(withPush, withPull)
}

func (r *Runner) RunPullOnce() error {
	_, processed, err := r.app.processOnePendingTask()
	if err != nil {
		return err
	}
	if !processed {
		r.app.logJSON("pull_noop", map[string]any{"consumer": r.app.cfg.Consumer})
	}
	return nil
}

func (r *Runner) RunPlay() error {
	return r.app.runPlay()
}

func (r *Runner) RunNamedTestScenario(scenario string) error {
	r.app.cfg.TestScenario = scenario
	return r.app.runNamedTestScenario(scenario)
}

func SupportedTestScenarios() []string {
	return append([]string(nil), supportedTestScenarios...)
}
