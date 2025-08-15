# Changelog

## [1.3.0](https://github.com/nentgroup/viaplay-cli/compare/v1.2.2...v1.3.0) (2025-08-15)


### Features

* add support for org teams ([#37](https://github.com/nentgroup/viaplay-cli/issues/37)) ([964e2c1](https://github.com/nentgroup/viaplay-cli/commit/964e2c100e0327b8e1c66a76e3c0af1a221c2407))

## [1.2.2](https://github.com/nentgroup/viaplay-cli/compare/v1.2.1...v1.2.2) (2025-08-14)


### Documentation

* update documentation and add MIT license ([#35](https://github.com/nentgroup/viaplay-cli/issues/35)) ([a1c38fb](https://github.com/nentgroup/viaplay-cli/commit/a1c38fbf1ece3546cf3bb1b18cb6ac827056e9dc))

## [1.2.1](https://github.com/nentgroup/viaplay-cli/compare/v1.2.0...v1.2.1) (2025-08-14)


### Bug Fixes

* fix secrets parsing ([#33](https://github.com/nentgroup/viaplay-cli/issues/33)) ([48c20d4](https://github.com/nentgroup/viaplay-cli/commit/48c20d47abb5f8a789defe557738b94a7dc7a93a))

## [1.2.0](https://github.com/nentgroup/viaplay-cli/compare/v1.1.0...v1.2.0) (2025-08-13)


### Features

* add new template functions ([#31](https://github.com/nentgroup/viaplay-cli/issues/31)) ([4e92f1f](https://github.com/nentgroup/viaplay-cli/commit/4e92f1f7401e5eeaff1f5f42bec1590f7f3ce7f1))

## [1.1.0](https://github.com/nentgroup/viaplay-cli/compare/v1.0.0...v1.1.0) (2025-08-13)


### Features

* add topics support and automatic push on repo creation ([#29](https://github.com/nentgroup/viaplay-cli/issues/29)) ([0c8d1e5](https://github.com/nentgroup/viaplay-cli/commit/0c8d1e5128eea5390c3d92c9f7cbcc66a6f88c4d))

## [1.0.0](https://github.com/nentgroup/viaplay-cli/compare/v0.4.0...v1.0.0) (2025-08-13)


### ⚠ BREAKING CHANGES

* Templates must now use {{ }} syntax instead of {{{ }}}.
* Tags now use Name Space format, e.g., .Repo.Name instead of .RepoName.
* Update all templates accordingly.

### Dependencies

* **deps:** update actions/checkout action to v5 ([#25](https://github.com/nentgroup/viaplay-cli/issues/25)) ([1a37fab](https://github.com/nentgroup/viaplay-cli/commit/1a37fabaddb1b0c35bf5436eb1927dbf8e46bb3b))
* **deps:** update dependency go to 1.25 ([#27](https://github.com/nentgroup/viaplay-cli/issues/27)) ([b75c611](https://github.com/nentgroup/viaplay-cli/commit/b75c61120bbbaa891abc8cde63201a2f5904faf6))


### Code Refactoring

* change template delimiters and tag naming ([#28](https://github.com/nentgroup/viaplay-cli/issues/28)) ([8788e8e](https://github.com/nentgroup/viaplay-cli/commit/8788e8e3f0b79aa3599e950a14a5d8ab59854ffa))

## [0.4.0](https://github.com/nentgroup/viaplay-cli/compare/v0.3.1...v0.4.0) (2025-08-10)


### Features

* add clean up feature ([50fddf4](https://github.com/nentgroup/viaplay-cli/commit/50fddf4dc1e2fe92ac29e0a005b96843df1a8851))
* consolidate output data ([191f6a4](https://github.com/nentgroup/viaplay-cli/commit/191f6a4c3afd3f0ea46344dde13cece7a49cd108))


### Bug Fixes

* show always banner on help ([b37d4c3](https://github.com/nentgroup/viaplay-cli/commit/b37d4c30f8544cd76e6e8d50e9612421bd03ea96))

## [0.3.1](https://github.com/nentgroup/viaplay-cli/compare/v0.3.0...v0.3.1) (2025-08-08)


### Bug Fixes

* update all references to old repo name ([a53abd8](https://github.com/nentgroup/viaplay-cli/commit/a53abd8919c368f638a8006667848f5f02be49b4))

## [0.3.0](https://github.com/nentgroup/viaplay-cli/compare/v0.2.1...v0.3.0) (2025-08-08)


### Features

* implement hooks feature ([#20](https://github.com/nentgroup/viaplay-cli/issues/20)) ([e6d1b91](https://github.com/nentgroup/viaplay-cli/commit/e6d1b91a2123a84ce6dca5e50281cb2ad9bc9b52))


### Dependencies

* **deps:** update dependency font-awesome to v7 ([#10](https://github.com/nentgroup/viaplay-cli/issues/10)) ([ccb66bd](https://github.com/nentgroup/viaplay-cli/commit/ccb66bdc6339422b029a10a7460e4a628913dfda))
* **deps:** update module golang.org/x/crypto to v0.41.0 ([#22](https://github.com/nentgroup/viaplay-cli/issues/22)) ([6058acb](https://github.com/nentgroup/viaplay-cli/commit/6058acb17243ef6abc573b0b15020870975f020d))

## [0.2.1](https://github.com/nentgroup/viaplay-cli/compare/v0.2.0...v0.2.1) (2025-08-06)


### Dependencies

* **deps:** update dependency font-awesome to v6.7.2 ([#7](https://github.com/nentgroup/viaplay-cli/issues/7)) ([65abfa6](https://github.com/nentgroup/viaplay-cli/commit/65abfa6928e3a492f6ba7bb50b01de63a9b9bca8))
* **deps:** update module golang.org/x/crypto to v0.35.0 [security] ([#2](https://github.com/nentgroup/viaplay-cli/issues/2)) ([2e1ae1f](https://github.com/nentgroup/viaplay-cli/commit/2e1ae1f80f0382d52f551e213139f1b81c488b25))
* **deps:** update module golang.org/x/crypto to v0.40.0 ([#18](https://github.com/nentgroup/viaplay-cli/issues/18)) ([39895c6](https://github.com/nentgroup/viaplay-cli/commit/39895c62f69b359e556639fbc13e3e68944ff023))
* **deps:** update module golang.org/x/oauth2 to v0.27.0 [security] ([#3](https://github.com/nentgroup/viaplay-cli/issues/3)) ([afd0df8](https://github.com/nentgroup/viaplay-cli/commit/afd0df8418e43014044c678f15dce2205a4cf304))
* **deps:** update module golang.org/x/oauth2 to v0.30.0 ([#19](https://github.com/nentgroup/viaplay-cli/issues/19)) ([000428b](https://github.com/nentgroup/viaplay-cli/commit/000428bb33aebfb6ebfc16f7bca9576fabf99043))

## [0.2.0](https://github.com/nentgroup/viaplay-cli/compare/v0.1.8...v0.2.0) (2025-08-04)


### Features

* add binary name flag ([00c4773](https://github.com/nentgroup/viaplay-cli/commit/00c4773a0d6af1bf01c818fd81f077c9cf42069a))

## [0.1.8](https://github.com/nentgroup/viaplay-cli/compare/v0.1.7...v0.1.8) (2025-08-04)


### Bug Fixes

* expand tilde ~ to homemapa when handling config files ([2535083](https://github.com/nentgroup/viaplay-cli/commit/2535083e6107281cf0a919834b63e18a317b3d57))

## [0.1.7](https://github.com/nentgroup/viaplay-cli/compare/v0.1.6...v0.1.7) (2025-08-04)


### Bug Fixes

* fix typo when blueprinting config files ([06f54c4](https://github.com/nentgroup/viaplay-cli/commit/06f54c4970baa9de031cfb0c09c9c8c96ee802be))

## [0.1.6](https://github.com/nentgroup/viaplay-cli/compare/v0.1.5...v0.1.6) (2025-08-04)


### Bug Fixes

* fix gh login bug caused by serialisation issues ([602a381](https://github.com/nentgroup/viaplay-cli/commit/602a3815ccea8179402fc3e03269f77384390e0d))

## [0.1.5](https://github.com/nentgroup/viaplay-cli/compare/v0.1.4...v0.1.5) (2025-08-04)


### Bug Fixes

* update goreleaser config ([cd50b3b](https://github.com/nentgroup/viaplay-cli/commit/cd50b3ba721d9604b6a767c3887b606892a0e21f))

## [0.1.4](https://github.com/nentgroup/viaplay-cli/compare/v0.1.3...v0.1.4) (2025-08-04)


### Bug Fixes

* add uninstall uninstall formula ([2fcb9fe](https://github.com/nentgroup/viaplay-cli/commit/2fcb9fe6d161eb5ce904f119c0464c93ea0819cb))

## [0.1.2](https://github.com/nentgroup/viaplay-cli/compare/v0.1.1...v0.1.2) (2025-08-04)


### Bug Fixes

* update brew formula ([7568187](https://github.com/nentgroup/viaplay-cli/commit/756818797b952f89a97c55b31ed86ad9ff7156d0))


### Documentation

* update readme ([049229f](https://github.com/nentgroup/viaplay-cli/commit/049229f8c848ea298e6432c2e4b4cbb7da6855d4))

## [0.1.1](https://github.com/nentgroup/viaplay-cli/compare/v0.1.0...v0.1.1) (2025-08-04)


### Bug Fixes

* rename client_id variable ([5545e39](https://github.com/nentgroup/viaplay-cli/commit/5545e39a2f98657666fea8877d137149d2445523))
* update release tags ([57bcc71](https://github.com/nentgroup/viaplay-cli/commit/57bcc71b34ae5d1839073483f1148c5229cd64b9))

## 0.1.0 (2025-08-04)


### Features

* initial commit ([281a451](https://github.com/nentgroup/viaplay-cli/commit/281a45143c444ce687fdad9e46ec17ac9425a28f))


### Bug Fixes

* update release flow ([339d20e](https://github.com/nentgroup/viaplay-cli/commit/339d20e1eb2bc63efc10f1bf4604c3fd1e3e5dbe))
