# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Changed SMTP_USERNAME and SMTP_PASSWORD to SENSU_EMAIL_SMTP_USERNAME and SENSU_EMAIL_SMTP_PASSWORD for consistency
- Changed some error returns to be more consistent

### Added
- Added support for annotations for certain variables

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
