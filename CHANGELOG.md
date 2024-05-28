# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Updated to use Go 1.22.x

## [1.2.2] - 2021-08-19

### Fixed
- Add StringLines helper template function and README example to make it possible to correctly format multi-line check output 
for html email body using golang template language.

## [1.2.1] - 2021-04-22
### Fixed
- Updated to latest sensu-plugin-sdk to workaround issue with agent config file forcing downcasing of agent annotations.
  Now downcased annotatation path for email handler configuration will be checked as a fallback if preferred camelcased keys are not present. 

## [1.2.0] - 2021-04-16

### Changed
- Added sprig package to provide enhanced template functioning capabilities

## [1.1.0] - 2021-03-18

### Changed
- Use net.JoinHostPort instead of fmt.Sprintf for Host:Port to make semgrep happy
- Updated Sensu Plugin SDK version (0.12)

## [1.0.0] - 2020-10-30

### Changed
- More template information in the README
- Q1 '21 handler maintenance:
  - Updated GitHub Actions: Added Lint action
  - Updated build to Go 1.14
  - Added Secret: true to SMTP password
  - Updated Bonsai to fix Windows amd64 build
  - Added output log line for email sent
  - Updated modules (go get -u && go mod tidy)
  - README updates

## [0.9.0] - 2020-10-30

### Added
- Add template function to allow expansion of event ID

### Changed
- Updated included sample event to include event ID

## [0.8.1] - 2020-10-22

### Changed
- Removed darwin 386 from bonsai

## [0.8.0] - 2020-10-22

### Changed
- Updated Sensu SDK version
- Updated Go version in go.mod
- Updated sample event.json to include timestamp attribute
- Added pull_request to test GitHub Action

## [0.7.0] - 2020-07-22

### Added
- Add template function to allow formatting of event timestamps

## [0.6.0] - 2020-05-27

### Changed
- Remove replacing newlines with HTML line breaks in HTML

## [0.5.2] - 2020-05-27

### Changed
- Updated sensu-plugin-sdk to v0.6.2

## [0.5.1] - 2020-03-25

### Changed
- Updated to use Go 1.14.x

## [0.5.0] - 2020-03-25

### Changed
- Use html/template for html email body
- Use a slice for -t (toEmail) to accept multiple recipients

## [0.4.1] - 2020-02-12

### Fixed
- Make goreleaser use SHA512

## [0.4.0] - 2020-02-12

### Fixed
- For html emails, sub &lt;br&gt; for \n in body

### Changed
- Fixed goreleaser deprecated archive to use archives
- Now depends on github.com/sensu-community/sensu-plugin-sdk@v0.6.0
- Now using Github Actions CI instead of Travis CI

### Added
- Added Date: header

## [0.3.0] - 2020-01-23

### Changed
- Deprecated the `insecure` flag. Use the combination of `--tlsSkipVerify`,
  `--authMethod none`, and `--smtpPort 25` for the same behavior.
- Deprecated the `enableLoginAuth` flag in favor of `--authMethod login`.

### Added
- Added `tlsSkipVerify` option to disable TLS certificate checks
- Added `authMethod` flag to switch between none, plain, and login auth

## [0.2.0] - 2019-07-23

### Changed
- Updated documentation to reflect loginauth and annotations

### Added
- Added loginauth auth mechanism
- Added Content-Type to support HTML emails
- Added plugin Keyspace so as to support configuraion overrides with annotations
- Added subject templating via argument/annotation

## [0.1.0] - 2019-03-04

### Changed
- Updated travis, goreleaser configurations.
- Updated license.

### Added
- Added option to allow for including check hook output in the email body
- Added option to allow for using insecure (port 25) email relays
- Added from address verification, RFC5322 From: header
- Added body template file
- Added environment variables for SMTP username and password

## [0.0.2] - 2018-12-12

### Added
- MIT license file

## [0.0.1] - 2018-12-12

### Added
- Initial release
