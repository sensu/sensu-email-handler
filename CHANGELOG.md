# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
