package main

import (
	"context"

	"dagger/kaza-presence-tracker-freebox/internal/dagger"
)

// Job: commit-msg
func (m *KazaPresenceTrackerFreebox) LintCommitMsg(ctx context.Context) (string, error) {
	return dag.Commitlint().Lint(m.Worktree, dagger.CommitlintLintOpts{Args: []string{"-l"}}).Stdout(ctx)
}

// Job: jsonfile
func (m *KazaPresenceTrackerFreebox) LintJsonFile(ctx context.Context) (string, error) {
	return dag.Jsonfile().Lint(ctx, m.Worktree)
}

// Job: gofmt
func (m *KazaPresenceTrackerFreebox) LintGofmt(ctx context.Context) ([]string, error) {
	return dag.Go(goVersion, m.Worktree).Fmt(ctx)
}

// Job: golang
func (m *KazaPresenceTrackerFreebox) LintGolang(ctx context.Context) (string, error) {
	return dag.Go(goVersion, m.Worktree).Lint(ctx)
}
