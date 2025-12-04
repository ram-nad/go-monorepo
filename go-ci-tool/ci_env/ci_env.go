// Package cienv cotains code for detecting CI environment
package cienv

import "os"

type CIEnvType string

const (
	None           CIEnvType = ""
	GitHubActions  CIEnvType = "GITHUB_ACTIONS"
	GiteaActions   CIEnvType = "GITEA_ACTIONS"
	Travis         CIEnvType = "TRAVIS"
	CircleCI       CIEnvType = "CIRCLECI"
	AppVeyor       CIEnvType = "APPVEYOR"
	GitLabCI       CIEnvType = "GITLAB_CI"
	BuildKite      CIEnvType = "BUILDKITE"
	Drone          CIEnvType = "DRONE"
	Codeship       CIEnvType = "CODESHIP"
	AzurePipelines CIEnvType = "AZURE_PIPELINES"
)

func GetCIEnvType() CIEnvType {
	if _, ok := os.LookupEnv("CI"); !ok {
		return None
	}

	if _, ok := os.LookupEnv("GITHUB_ACTIONS"); ok {
		return GitHubActions
	}

	if _, ok := os.LookupEnv("GITEA_ACTIONS"); ok {
		return GiteaActions
	}

	if _, ok := os.LookupEnv("TRAVIS"); ok {
		return Travis
	}

	if _, ok := os.LookupEnv("CIRCLECI"); ok {
		return CircleCI
	}

	if _, ok := os.LookupEnv("APPVEYOR"); ok {
		return AppVeyor
	}

	if _, ok := os.LookupEnv("GITLAB_CI"); ok {
		return GitLabCI
	}

	if _, ok := os.LookupEnv("BUILDKITE"); ok {
		return BuildKite
	}

	if _, ok := os.LookupEnv("DRONE"); ok {
		return Drone
	}

	if ciName, ok := os.LookupEnv("CI_NAME"); ok && ciName == "codeship" {
		return Codeship
	}

	if _, ok := os.LookupEnv("TF_BUILD"); ok {
		return AzurePipelines
	}

	return None
}

func IsCIEnvAndSupportsColor() bool {
	ciEnv := GetCIEnvType()

	if ciEnv == GitHubActions ||
		ciEnv == GiteaActions ||
		ciEnv == Travis ||
		ciEnv == CircleCI ||
		ciEnv == AppVeyor ||
		ciEnv == GitLabCI ||
		ciEnv == BuildKite ||
		ciEnv == Drone ||
		ciEnv == Codeship {
		return true
	}

	return false
}
