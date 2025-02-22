package gittag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Target creates a tag if needed from a local git repository, without pushing the tag
func (gt *GitTag) Target(source string, dryRun bool) (changed bool, err error) {
	changed, _, _, err = gt.target(source, dryRun)

	return changed, err
}

// TargetFromSCM creates and pushes a git tag based on the SCM configuration
func (gt *GitTag) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {
	if len(gt.spec.Path) > 0 {
		logrus.Warningf("Path setting value %q overridden by the scm configuration (value %q)",
			gt.spec.Path,
			scm.GetDirectory())
	}
	gt.spec.Path = scm.GetDirectory()

	changed, files, message, err = gt.target(source, dryRun)
	if err != nil {
		return changed, files, message, err
	}

	err = scm.PushTag(source)
	if err != nil {
		logrus.Errorf("Git push tag error: %s", err)
		return changed, files, message, err
	}
	logrus.Infof("%s The git tag %q was pushed successfully to the specified remote.", result.ATTENTION, source)
	return changed, files, message, err
}

func (gt *GitTag) target(source string, dryRun bool) (bool, []string, string, error) {
	// Ensure that a git message is present to annotate the tag to create
	if len(gt.spec.Message) == 0 {
		// absence of a message is not blocking: warn the user and continue
		gt.spec.Message = "Generated by updatecli"
		logrus.Warningf("No specified message for gittag target. Using default value %q", gt.spec.Message)
	}

	files := []string{}
	message := gt.spec.Message

	// Fail if a pattern is specified
	if gt.spec.VersionFilter.Pattern != "" {
		return false, files, message, fmt.Errorf("Target validation error: spec.versionfilter.pattern is not allowed for targets of type gittag.")
	}

	// Fail if the git tag resource cannot be validated
	err := gt.Validate()
	if err != nil {
		logrus.Errorln(err)
		return false, files, message, err
	}

	// Check if the provided tag (from source input value) already exists
	gt.spec.VersionFilter.Pattern = source
	tags, err := gitgeneric.Tags(gt.spec.Path)
	if err != nil {
		return false, files, message, err
	}
	gt.foundVersion, err = gt.spec.VersionFilter.Search(tags)
	if err != nil && err.Error() != fmt.Sprintf("No version found matching pattern %q", source) {
		// Something went wrong during the tag search.
		return false, files, message, err
	}
	if gt.foundVersion.ParsedVersion == source {
		// No error, but no change
		logrus.Printf("%s The Git Tag %q already exists, nothing else to do.",
			result.SUCCESS,
			source)
		return false, files, message, nil
	}

	// Otherwise proceed to create this new tag
	logrus.Printf("%s The git tag %q does not exist: creating it.", result.ATTENTION, source)

	if dryRun {
		// Dry run: no changes to apply.
		// Return early without creating tag but notify that a change should be made.
		return true, files, message, nil
	}

	changed, err := gitgeneric.NewTag(source, gt.spec.Message, gt.spec.Path)
	if err != nil {
		return changed, files, message, err
	}
	logrus.Printf("%s The git tag %q has been created.", result.ATTENTION, source)

	return changed, files, message, nil
}
