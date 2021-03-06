// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package server

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/mattermost/mattermost-mattermod/model"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/pkg/errors"
)

func (s *Server) addHacktoberfestLabel(ctx context.Context, pr *model.PullRequest) {
	if pr.State == model.StateClosed {
		return
	}

	// Ignore PRs not created in october
	if pr.CreatedAt.Month() != time.October {
		return
	}

	// Don't apply label if the contributors is a core committer
	if s.IsOrgMember(pr.Username) {
		return
	}

	_, _, err := s.GithubClient.Issues.AddLabelsToIssue(ctx, pr.RepoOwner, pr.RepoName, pr.Number, []string{"Hacktoberfest", "hacktoberfest-accepted"})
	if err != nil {
		mlog.Error("Error applying Hacktoberfest labels", mlog.Err(err), mlog.Int("PR", pr.Number), mlog.String("Repo", pr.RepoName))
		return
	}
}

func (s *Server) postPRWelcomeMessage(ctx context.Context, pr *model.PullRequest, claCommentNeeded bool) error {
	// Only post welcome Message for community member
	if s.IsOrgMember(pr.Username) {
		return nil
	}

	t, err := template.New("welcomeMessage").Parse(s.Config.PRWelcomeMessage)
	if err != nil {
		return errors.Wrap(err, "failed to render welcome message template")
	}

	var output bytes.Buffer
	data := map[string]interface{}{
		"CLACommentNeeded": claCommentNeeded,
		"Username":         "@" + pr.Username,
	}
	err = t.Execute(&output, data)
	if err != nil {
		return errors.Wrap(err, "could not execute welcome message template")
	}

	err = s.sendGitHubComment(ctx, pr.RepoOwner, pr.RepoName, pr.Number, output.String())
	if err != nil {
		return errors.Wrap(err, "failed to send welcome message")
	}

	return nil
}
