# Changelog

## [3.4.0](https://github.com/nentgroup/viaplay-cli/compare/v3.3.2...v3.4.0) (2026-09-17)


### Features

* add support for remote template hooks ([#213](https://github.com/nentgroup/viaplay-cli/issues/213)) ([a575ba2](https://github.com/nentgroup/viaplay-cli/commit/a575ba2a2ef8e821559d2a8462d9417ea7422249))

## [3.3.2](https://github.com/nentgroup/viaplay-cli/compare/v3.3.1...v3.3.2) (2026-09-16)


### Bug Fixes

* make functionmap available for files ([#211](https://github.com/nentgroup/viaplay-cli/issues/211)) ([8aa74fa](https://github.com/nentgroup/viaplay-cli/commit/8aa74fafb3ce041ec8be3c3393517c5bfbae51c9))

## [3.3.1](https://github.com/nentgroup/viaplay-cli/compare/v3.3.0...v3.3.1) (2026-09-13)


### Bug Fixes

* **homebrew:** skip macOS-only xattr postflight step on Linux ([8d1252f](https://github.com/nentgroup/viaplay-cli/commit/8d1252f687099e0ffc1e96c7776f5a5fc5795176))


### Documentation

* update documentation ([126d9ac](https://github.com/nentgroup/viaplay-cli/commit/126d9ac09ed9ab1932d504d3d8ca2c9f41eada69))

## [3.3.0](https://github.com/nentgroup/viaplay-cli/compare/v3.2.0...v3.3.0) (2026-09-13)


### Features

* prepare for public release ([#203](https://github.com/nentgroup/viaplay-cli/issues/203)) ([ffb89a8](https://github.com/nentgroup/viaplay-cli/commit/ffb89a8b2f5175a1f01a260651d852342bf374e7))

## [3.2.0](https://github.com/nentgroup/viaplay-cli/compare/v3.1.0...v3.2.0) (2026-09-12)


### Features

* **template:** add shared string and ID helpers ([#202](https://github.com/nentgroup/viaplay-cli/issues/202)) ([59c3453](https://github.com/nentgroup/viaplay-cli/commit/59c3453dc6a4006391bebbb19430fb128ff167c2))


### Dependencies

* **deps:** update module golang.org/x/crypto to v0.57.0 ([#196](https://github.com/nentgroup/viaplay-cli/issues/196)) ([292dc47](https://github.com/nentgroup/viaplay-cli/commit/292dc473ac812b183188a4d4e5ef7f1d108e8ab3))
* **deps:** update module golang.org/x/oauth2 to v0.37.0 ([#193](https://github.com/nentgroup/viaplay-cli/issues/193)) ([df7f9dc](https://github.com/nentgroup/viaplay-cli/commit/df7f9dcd5085b4558fdee5762a7295ecdd694846))

## [3.1.0](https://github.com/nentgroup/viaplay-cli/compare/v3.0.0...v3.1.0) (2026-09-09)


### Features

* **template:** add authors metadata to manifest  ([#200](https://github.com/nentgroup/viaplay-cli/issues/200)) ([5572177](https://github.com/nentgroup/viaplay-cli/commit/5572177bfe752cad3a94bee7e712b56a8f371d22))


### Bug Fixes

* silent team default and hook order ([#198](https://github.com/nentgroup/viaplay-cli/issues/198)) ([21411ad](https://github.com/nentgroup/viaplay-cli/commit/21411ad6c904511f6a1578b2c69a8eb810f329ae))

## [3.0.0](https://github.com/nentgroup/viaplay-cli/compare/v2.9.0...v3.0.0) (2026-09-08)


### ⚠ BREAKING CHANGES

* `vip template remove` now deletes the cached template clone by default (use --keep-cache to preserve it) and resolves team-first when a team is configured, falling back to personal config only if the team has none. Manifest structure is now validated on template add/inspect/test/project create; previously-malformed manifests that used to silently work may now fail.

### Features

* add skills command and extend template ([#194](https://github.com/nentgroup/viaplay-cli/issues/194)) ([b50608a](https://github.com/nentgroup/viaplay-cli/commit/b50608a9a901b13ce7573e6c68cbceda46c55765))

## [2.9.0](https://github.com/nentgroup/viaplay-cli/compare/v2.8.2...v2.9.0) (2026-09-06)


### Features

* add interactive templates ([#192](https://github.com/nentgroup/viaplay-cli/issues/192)) ([1b67e94](https://github.com/nentgroup/viaplay-cli/commit/1b67e942670ade8d4632101ab30cd7dc33738069))


### Dependencies

* **deps:** update module github.com/google/go-github/v74 to v91 ([#189](https://github.com/nentgroup/viaplay-cli/issues/189)) ([f239d49](https://github.com/nentgroup/viaplay-cli/commit/f239d4914b6a696b242d0f00e584da43686d9e20))
* **deps:** update module golang.org/x/crypto to v0.56.0 ([#188](https://github.com/nentgroup/viaplay-cli/issues/188)) ([82bfaca](https://github.com/nentgroup/viaplay-cli/commit/82bfaca20abbdf1fe7484db281ea31eb7c2280fa))

## [2.8.2](https://github.com/nentgroup/viaplay-cli/compare/v2.8.1...v2.8.2) (2026-08-21)


### Dependencies

* upgrade all modules ([078fdd6](https://github.com/nentgroup/viaplay-cli/commit/078fdd60b2cddc9203892a3cc461d60f7072b2e5))

## [2.8.1](https://github.com/nentgroup/viaplay-cli/compare/v2.8.0...v2.8.1) (2026-08-21)


### Documentation

* upgrade docsify ([81a337d](https://github.com/nentgroup/viaplay-cli/commit/81a337d9c1b8ea847edc2630e58cad535bbccd8b))


### Dependencies

* **deps:** update google/go-github/v74 to v90 ([#180](https://github.com/nentgroup/viaplay-cli/issues/180)) ([7181296](https://github.com/nentgroup/viaplay-cli/commit/7181296634932dca7a1512011d0fc20cdd14d7af))

## [2.8.0](https://github.com/nentgroup/viaplay-cli/compare/v2.7.0...v2.8.0) (2026-08-21)


### Features

* add hooks command ([ff40d19](https://github.com/nentgroup/viaplay-cli/commit/ff40d19f3cb03109dfe4e443954ca0d78e1eae9e))
* add support for remote shared configs ([3349548](https://github.com/nentgroup/viaplay-cli/commit/3349548faf6d03bd8abbfde28962692d191079c7))
* enhance config command ergonomics ([975ee24](https://github.com/nentgroup/viaplay-cli/commit/975ee2480ef825da3769d523f489d157734e040a))
* enhance template command ergonomics ([fcdf4b4](https://github.com/nentgroup/viaplay-cli/commit/fcdf4b46582af04b9d8dcfe8ca5616de7ddafe4b))
* refactor hooks discovery ([5f819f5](https://github.com/nentgroup/viaplay-cli/commit/5f819f56fe5ba0a703d949278698159dddfe2501))


### Bug Fixes

* correct apply --only flag ([69cb9c0](https://github.com/nentgroup/viaplay-cli/commit/69cb9c010d983804375891f721b6e3a366bea309))
* refactor shared hooks ([9b716c5](https://github.com/nentgroup/viaplay-cli/commit/9b716c5ad5fb74339e380c661f9c9b5f9b41fbf9))

## [2.7.0](https://github.com/nentgroup/viaplay-cli/compare/v2.6.0...v2.7.0) (2026-08-20)


### Features

* apply secrets to existing repos ([eb071aa](https://github.com/nentgroup/viaplay-cli/commit/eb071aa47e44ea2b63aa7fe2704d3829e334bb4a))


### Documentation

* clean documentation ([#148](https://github.com/nentgroup/viaplay-cli/issues/148)) ([1608a05](https://github.com/nentgroup/viaplay-cli/commit/1608a05d002d9259f1af98852792692a0a730d87))
* update readme and documentation ([#146](https://github.com/nentgroup/viaplay-cli/issues/146)) ([5122e98](https://github.com/nentgroup/viaplay-cli/commit/5122e98823b9d462b2a5b27d4428701440dc549f))


### Dependencies

* **deps:** update actions/checkout action to v7 ([#169](https://github.com/nentgroup/viaplay-cli/issues/169)) ([82952de](https://github.com/nentgroup/viaplay-cli/commit/82952de240fd11963340e36ac7402e985ab9d11d))
* **deps:** update actions/labeler action to v7 ([#177](https://github.com/nentgroup/viaplay-cli/issues/177)) ([35f46dc](https://github.com/nentgroup/viaplay-cli/commit/35f46dc2be816a1cd5d8db42c3c082ff87c3b60c))
* **deps:** update actions/setup-go action to v7 ([#176](https://github.com/nentgroup/viaplay-cli/issues/176)) ([f1ff33d](https://github.com/nentgroup/viaplay-cli/commit/f1ff33d5d9367b979468fd1ccf8cdb21a5ceccec))
* **deps:** update actions/upload-artifact action to v7 ([#134](https://github.com/nentgroup/viaplay-cli/issues/134)) ([92b2ffc](https://github.com/nentgroup/viaplay-cli/commit/92b2ffc5f37a538edea71ea77f15a596ee70276f))
* **deps:** update dependency font-awesome to v7.3.0 ([#175](https://github.com/nentgroup/viaplay-cli/issues/175)) ([c106989](https://github.com/nentgroup/viaplay-cli/commit/c106989bd3f26225928874725ce28647c5c9b28c))
* **deps:** update dependency font-awesome to v7.3.1 ([#182](https://github.com/nentgroup/viaplay-cli/issues/182)) ([b05a746](https://github.com/nentgroup/viaplay-cli/commit/b05a746be15c25e77a115c36128831910be6a0e3))
* **deps:** update dependency go to 1.26 ([#130](https://github.com/nentgroup/viaplay-cli/issues/130)) ([d4d6ec4](https://github.com/nentgroup/viaplay-cli/commit/d4d6ec40bb7020655d455b6d5c169696b9070de6))
* **deps:** update dependency go to 1.27 ([#183](https://github.com/nentgroup/viaplay-cli/issues/183)) ([12edc5b](https://github.com/nentgroup/viaplay-cli/commit/12edc5bae4d18b21ef2d8d4363aa829b976ce5ea))
* **deps:** update googleapis/release-please-action action to v5 ([#154](https://github.com/nentgroup/viaplay-cli/issues/154)) ([e6e291c](https://github.com/nentgroup/viaplay-cli/commit/e6e291cf4a8c08f32529207f43660098aed4226a))
* **deps:** update goreleaser/goreleaser-action action to v7 ([#133](https://github.com/nentgroup/viaplay-cli/issues/133)) ([b52fb81](https://github.com/nentgroup/viaplay-cli/commit/b52fb810bdec98b5d92e7bbbef73785b8d8fbf1b))
* **deps:** update module github.com/charmbracelet/bubbles to v1 ([#128](https://github.com/nentgroup/viaplay-cli/issues/128)) ([f15c5f5](https://github.com/nentgroup/viaplay-cli/commit/f15c5f52a2bffd5d27f4a413762c996b7c9ca91b))
* **deps:** update module github.com/fatih/color to v1.19.0 ([#144](https://github.com/nentgroup/viaplay-cli/issues/144)) ([bda77fa](https://github.com/nentgroup/viaplay-cli/commit/bda77fa524df59464fccf395fbb93819e2470225))
* **deps:** update module github.com/google/go-github/v74 to v83 ([#131](https://github.com/nentgroup/viaplay-cli/issues/131)) ([9eae730](https://github.com/nentgroup/viaplay-cli/commit/9eae730d2a9427294425d34ee808543a333f242d))
* **deps:** update module github.com/google/go-github/v74 to v84 ([#138](https://github.com/nentgroup/viaplay-cli/issues/138)) ([712e732](https://github.com/nentgroup/viaplay-cli/commit/712e7327ed7839a67dc4bb6cf931875acfeced90))
* **deps:** update module github.com/google/go-github/v74 to v85 ([#151](https://github.com/nentgroup/viaplay-cli/issues/151)) ([c181b95](https://github.com/nentgroup/viaplay-cli/commit/c181b95c684eb97149d0bcabe903188344f1a57e))
* **deps:** update module github.com/google/go-github/v74 to v86 ([#155](https://github.com/nentgroup/viaplay-cli/issues/155)) ([f61e3b2](https://github.com/nentgroup/viaplay-cli/commit/f61e3b23008577dc9ce52ddd0d50361de30737cf))
* **deps:** update module github.com/google/go-github/v74 to v87 ([#160](https://github.com/nentgroup/viaplay-cli/issues/160)) ([9946d52](https://github.com/nentgroup/viaplay-cli/commit/9946d52ef859d5493315a537911bd25297be4ba7))
* **deps:** update module github.com/google/go-github/v74 to v88 ([#163](https://github.com/nentgroup/viaplay-cli/issues/163)) ([4611508](https://github.com/nentgroup/viaplay-cli/commit/4611508c445813d9a1b41053492905da52a49a65))
* **deps:** update module github.com/google/go-github/v74 to v89 ([#170](https://github.com/nentgroup/viaplay-cli/issues/170)) ([72013c8](https://github.com/nentgroup/viaplay-cli/commit/72013c825dced3cf4a18f8e629c08eb70c87b426))
* **deps:** update module github.com/google/go-github/v74 to v90 ([#178](https://github.com/nentgroup/viaplay-cli/issues/178)) ([08a037f](https://github.com/nentgroup/viaplay-cli/commit/08a037f92e924de8ee79f496872ce311c126543e))
* **deps:** update module github.com/google/go-github/v83 to v84 ([#139](https://github.com/nentgroup/viaplay-cli/issues/139)) ([f1bb0aa](https://github.com/nentgroup/viaplay-cli/commit/f1bb0aa39dc24ed1bc4e16a1aee77812541cbc0d))
* **deps:** update module github.com/google/go-github/v84 to v85 ([#152](https://github.com/nentgroup/viaplay-cli/issues/152)) ([acdbc8a](https://github.com/nentgroup/viaplay-cli/commit/acdbc8a9de5b1a2575db8ea74a88efbb1c525133))
* **deps:** update module github.com/google/go-github/v85 to v86 ([#156](https://github.com/nentgroup/viaplay-cli/issues/156)) ([32dc5ae](https://github.com/nentgroup/viaplay-cli/commit/32dc5ae91568fbfe7d2032d3bafcae43b64821e9))
* **deps:** update module github.com/google/go-github/v86 to v87 ([#161](https://github.com/nentgroup/viaplay-cli/issues/161)) ([ea47d53](https://github.com/nentgroup/viaplay-cli/commit/ea47d53e6dbbf9146b1864a1a17e5389f1b7956a))
* **deps:** update module github.com/google/go-github/v87 to v88 ([#164](https://github.com/nentgroup/viaplay-cli/issues/164)) ([1844473](https://github.com/nentgroup/viaplay-cli/commit/1844473bc1d5a2830fb0bf30255f5687754714eb))
* **deps:** update module github.com/google/go-github/v88 to v89 ([#171](https://github.com/nentgroup/viaplay-cli/issues/171)) ([0f935da](https://github.com/nentgroup/viaplay-cli/commit/0f935dab9073e51715a6336e52c9d0e891800dcc))
* **deps:** update module github.com/google/go-github/v89 to v90 ([#179](https://github.com/nentgroup/viaplay-cli/issues/179)) ([322e7e6](https://github.com/nentgroup/viaplay-cli/commit/322e7e65f283e975fd0de23ded137e3a565904e2))
* **deps:** update module github.com/zalando/go-keyring to v0.2.7 ([#145](https://github.com/nentgroup/viaplay-cli/issues/145)) ([b5b5671](https://github.com/nentgroup/viaplay-cli/commit/b5b567177e2fce58f1edd68682c22551d133185e))
* **deps:** update module github.com/zalando/go-keyring to v0.2.8 ([#147](https://github.com/nentgroup/viaplay-cli/issues/147)) ([a084e2b](https://github.com/nentgroup/viaplay-cli/commit/a084e2b021b41a62cf47dad55bbfb9f4e0819568))
* **deps:** update module golang.org/x/crypto to v0.49.0 ([#143](https://github.com/nentgroup/viaplay-cli/issues/143)) ([a69c8ba](https://github.com/nentgroup/viaplay-cli/commit/a69c8ba23f7de21cceefddbf2bf34998c6c60585))
* **deps:** update module golang.org/x/crypto to v0.50.0 ([#150](https://github.com/nentgroup/viaplay-cli/issues/150)) ([ecb45d9](https://github.com/nentgroup/viaplay-cli/commit/ecb45d94ee1a2270ab6cbf19897f25d6617a41bb))
* **deps:** update module golang.org/x/crypto to v0.51.0 ([#159](https://github.com/nentgroup/viaplay-cli/issues/159)) ([9446350](https://github.com/nentgroup/viaplay-cli/commit/94463506745b064ffcb7fd5d9503fc71ceee75e6))
* **deps:** update module golang.org/x/crypto to v0.52.0 ([#166](https://github.com/nentgroup/viaplay-cli/issues/166)) ([b186b70](https://github.com/nentgroup/viaplay-cli/commit/b186b702b9bb974cb477704d0fb5a697d53347ba))
* **deps:** update module golang.org/x/crypto to v0.53.0 ([#167](https://github.com/nentgroup/viaplay-cli/issues/167)) ([61f8c3f](https://github.com/nentgroup/viaplay-cli/commit/61f8c3f2d274590b995970a22b7ee7a3c5ff5cd7))
* **deps:** update module golang.org/x/crypto to v0.54.0 ([#173](https://github.com/nentgroup/viaplay-cli/issues/173)) ([da900a8](https://github.com/nentgroup/viaplay-cli/commit/da900a855dc4926ffa4e1ea7542cbbe4000a47da))
* **deps:** update module golang.org/x/crypto to v0.55.0 ([#181](https://github.com/nentgroup/viaplay-cli/issues/181)) ([74cda12](https://github.com/nentgroup/viaplay-cli/commit/74cda12d555b2788b46faacf14374790cfbeaef9))
* **deps:** update module golang.org/x/oauth2 to v0.36.0 ([#141](https://github.com/nentgroup/viaplay-cli/issues/141)) ([0101871](https://github.com/nentgroup/viaplay-cli/commit/0101871603ab38b15c813819ff2d87a58825d822))
* **deps:** update module golang.org/x/term to v0.41.0 ([#142](https://github.com/nentgroup/viaplay-cli/issues/142)) ([d230f60](https://github.com/nentgroup/viaplay-cli/commit/d230f605102973956219ade77398a946ceafa703))
* **deps:** update module golang.org/x/term to v0.42.0 ([#149](https://github.com/nentgroup/viaplay-cli/issues/149)) ([53cccff](https://github.com/nentgroup/viaplay-cli/commit/53cccff25fa825bf6a5465fc38cbf64455840573))
* **deps:** update module golang.org/x/term to v0.43.0 ([#158](https://github.com/nentgroup/viaplay-cli/issues/158)) ([0864d03](https://github.com/nentgroup/viaplay-cli/commit/0864d0383bb4884cfff38a0747032ab72c6429e1))

## [2.6.0](https://github.com/nentgroup/viaplay-cli/compare/v2.5.0...v2.6.0) (2026-02-09)


### Features

* add support for local repo ([#126](https://github.com/nentgroup/viaplay-cli/issues/126)) ([eddf09d](https://github.com/nentgroup/viaplay-cli/commit/eddf09dd7f575d0880ef9c73d75ed6804797847a))


### Dependencies

* **deps:** update dependency go to v1.25.6 ([#122](https://github.com/nentgroup/viaplay-cli/issues/122)) ([0000d5f](https://github.com/nentgroup/viaplay-cli/commit/0000d5f68b41950f6aa6a7bb21bf16e5dcde40d8))
* **deps:** update dependency go to v1.25.7 ([#123](https://github.com/nentgroup/viaplay-cli/issues/123)) ([6128649](https://github.com/nentgroup/viaplay-cli/commit/61286493d4bc756b94d8478fc27aa39f48af1de1))
* **deps:** update module github.com/charmbracelet/bubbles to v0.21.1 ([#121](https://github.com/nentgroup/viaplay-cli/issues/121)) ([19167b8](https://github.com/nentgroup/viaplay-cli/commit/19167b8a1d7c8d0221e0efb621ad390ecebd52cb))
* **deps:** update module github.com/google/go-github/v74 to v82 ([#117](https://github.com/nentgroup/viaplay-cli/issues/117)) ([19ade1a](https://github.com/nentgroup/viaplay-cli/commit/19ade1af03de45445bc7e51510c4e624c99faff9))
* **deps:** update module github.com/google/go-github/v81 to v82 ([#118](https://github.com/nentgroup/viaplay-cli/issues/118)) ([ac6dddd](https://github.com/nentgroup/viaplay-cli/commit/ac6ddddc5e383ae60af6f715322728bbdf662203))
* **deps:** update module golang.org/x/crypto to v0.48.0 ([#127](https://github.com/nentgroup/viaplay-cli/issues/127)) ([51459bd](https://github.com/nentgroup/viaplay-cli/commit/51459bd490880b45d0a824a07aa5ae04282645e4))
* **deps:** update module golang.org/x/oauth2 to v0.35.0 ([#124](https://github.com/nentgroup/viaplay-cli/issues/124)) ([b996bad](https://github.com/nentgroup/viaplay-cli/commit/b996bad5af64462d9aa2902bf857d9419e239049))
* **deps:** update module golang.org/x/term to v0.40.0 ([#125](https://github.com/nentgroup/viaplay-cli/issues/125)) ([1ff8f51](https://github.com/nentgroup/viaplay-cli/commit/1ff8f51ff46a924e7189292f6e9e642f4a31bd3c))

## [2.5.0](https://github.com/nentgroup/viaplay-cli/compare/v2.4.2...v2.5.0) (2026-01-18)


### Features

* refactor cache list output ([#115](https://github.com/nentgroup/viaplay-cli/issues/115)) ([6484cd2](https://github.com/nentgroup/viaplay-cli/commit/6484cd2743a68600056faaaaf2deca88c0e08e45))

## [2.4.2](https://github.com/nentgroup/viaplay-cli/compare/v2.4.1...v2.4.2) (2026-01-12)


### Dependencies

* **deps:** update module golang.org/x/crypto to v0.47.0 ([#114](https://github.com/nentgroup/viaplay-cli/issues/114)) ([61bd9ba](https://github.com/nentgroup/viaplay-cli/commit/61bd9ba52cf70aac37bc85aaa18a6e330f68cedd))
* **deps:** update module golang.org/x/term to v0.39.0 ([#112](https://github.com/nentgroup/viaplay-cli/issues/112)) ([cbf7a55](https://github.com/nentgroup/viaplay-cli/commit/cbf7a55e9d2e0d09a46c02d02c7d8aaaa5aeb2eb))

## [2.4.1](https://github.com/nentgroup/viaplay-cli/compare/v2.4.0...v2.4.1) (2026-01-09)


### Documentation

* update dark svg icons ([#110](https://github.com/nentgroup/viaplay-cli/issues/110)) ([091d2eb](https://github.com/nentgroup/viaplay-cli/commit/091d2eb79530466f7842c84c735c9a1643b9b976))

## [2.4.0](https://github.com/nentgroup/viaplay-cli/compare/v2.3.2...v2.4.0) (2026-01-09)


### Features

* add support for .raw files ([#109](https://github.com/nentgroup/viaplay-cli/issues/109)) ([776bb5f](https://github.com/nentgroup/viaplay-cli/commit/776bb5ff5e80dda91a95d81449830b5760c87173))


### Dependencies

* **deps:** update actions/checkout action to v6 ([#97](https://github.com/nentgroup/viaplay-cli/issues/97)) ([10733f5](https://github.com/nentgroup/viaplay-cli/commit/10733f52d59c67ab14b0b692c56529ae012be9c6))
* **deps:** update actions/upload-artifact action to v5 ([#82](https://github.com/nentgroup/viaplay-cli/issues/82)) ([48b55ee](https://github.com/nentgroup/viaplay-cli/commit/48b55eeeaad556a165716ff2651390517217dcac))
* **deps:** update actions/upload-artifact action to v6 ([#105](https://github.com/nentgroup/viaplay-cli/issues/105)) ([2dad026](https://github.com/nentgroup/viaplay-cli/commit/2dad026c96d4a719710509a431855fd99600ab32))
* **deps:** update module github.com/charmbracelet/bubbletea to v1.3.10 ([#73](https://github.com/nentgroup/viaplay-cli/issues/73)) ([ae8b2a2](https://github.com/nentgroup/viaplay-cli/commit/ae8b2a2d6b4a4feb80e45d92f51eaf6c297ee909))
* **deps:** update module github.com/charmbracelet/bubbletea to v1.3.7 ([#64](https://github.com/nentgroup/viaplay-cli/issues/64)) ([b7e6d7f](https://github.com/nentgroup/viaplay-cli/commit/b7e6d7f2b11c16eb6d13f474d2033fcbe7ae0b20))
* **deps:** update module github.com/charmbracelet/bubbletea to v1.3.8 ([#70](https://github.com/nentgroup/viaplay-cli/issues/70)) ([9fa9408](https://github.com/nentgroup/viaplay-cli/commit/9fa9408aa9b7554f90164cdad24e6be2422ff2a8))
* **deps:** update module github.com/charmbracelet/bubbletea to v1.3.9 ([#72](https://github.com/nentgroup/viaplay-cli/issues/72)) ([223b85f](https://github.com/nentgroup/viaplay-cli/commit/223b85feb8b9abbf6f4f7e431bf731ec1f83e910))
* **deps:** update module github.com/google/go-github/v74 to v75 ([#74](https://github.com/nentgroup/viaplay-cli/issues/74)) ([c0775ef](https://github.com/nentgroup/viaplay-cli/commit/c0775ef40109678c40e2dde6acb03f6b93ca46bc))
* **deps:** update module github.com/google/go-github/v74 to v76 ([#79](https://github.com/nentgroup/viaplay-cli/issues/79)) ([f9d7138](https://github.com/nentgroup/viaplay-cli/commit/f9d71385815cf4fff990e7390277ab9a0a4c3fd9))
* **deps:** update module github.com/google/go-github/v74 to v77 ([#83](https://github.com/nentgroup/viaplay-cli/issues/83)) ([91cde4f](https://github.com/nentgroup/viaplay-cli/commit/91cde4f8b57a7dcae2c2aaafeb96f9015603ccad))
* **deps:** update module github.com/google/go-github/v74 to v78 ([#87](https://github.com/nentgroup/viaplay-cli/issues/87)) ([dae8ef3](https://github.com/nentgroup/viaplay-cli/commit/dae8ef3339d82067f2c2e7e2dd59a8cec707ea84))
* **deps:** update module github.com/google/go-github/v74 to v79 ([#92](https://github.com/nentgroup/viaplay-cli/issues/92)) ([eeddfcc](https://github.com/nentgroup/viaplay-cli/commit/eeddfcc16dbede02e5e306bc6c30107dde833f44))
* **deps:** update module github.com/google/go-github/v74 to v80 ([#99](https://github.com/nentgroup/viaplay-cli/issues/99)) ([b832515](https://github.com/nentgroup/viaplay-cli/commit/b832515749dc9442b451abeb0f677790210eea96))
* **deps:** update module github.com/google/go-github/v74 to v81 ([#106](https://github.com/nentgroup/viaplay-cli/issues/106)) ([bb0d875](https://github.com/nentgroup/viaplay-cli/commit/bb0d875c1d105fb090b4b6337a05cc31c44dc16b))
* **deps:** update module github.com/google/go-github/v75 to v76 ([#80](https://github.com/nentgroup/viaplay-cli/issues/80)) ([6f9e84e](https://github.com/nentgroup/viaplay-cli/commit/6f9e84e062241cdb6986895d79d684245e471f8e))
* **deps:** update module github.com/google/go-github/v76 to v77 ([#84](https://github.com/nentgroup/viaplay-cli/issues/84)) ([0861b17](https://github.com/nentgroup/viaplay-cli/commit/0861b17431e40bd809cfac972c0d82f2f8643d59))
* **deps:** update module github.com/google/go-github/v77 to v78 ([#88](https://github.com/nentgroup/viaplay-cli/issues/88)) ([73ee7c3](https://github.com/nentgroup/viaplay-cli/commit/73ee7c339e3fddbc04d863485d7e4f084a5c5ed9))
* **deps:** update module github.com/google/go-github/v78 to v79 ([#93](https://github.com/nentgroup/viaplay-cli/issues/93)) ([e7059ff](https://github.com/nentgroup/viaplay-cli/commit/e7059ff3778b82e30a3a3176f6e48a0d96731405))
* **deps:** update module github.com/google/go-github/v79 to v80 ([#100](https://github.com/nentgroup/viaplay-cli/issues/100)) ([3be62e7](https://github.com/nentgroup/viaplay-cli/commit/3be62e717ad6ed996b6d93f57c8bd51af3c9f8fe))
* **deps:** update module github.com/google/go-github/v80 to v81 ([#107](https://github.com/nentgroup/viaplay-cli/issues/107)) ([7549e79](https://github.com/nentgroup/viaplay-cli/commit/7549e79c7f014aa6a766d7e878b0b5b9ae4e2695))
* **deps:** update module github.com/spf13/afero to v1.15.0 ([#68](https://github.com/nentgroup/viaplay-cli/issues/68)) ([c71c4ac](https://github.com/nentgroup/viaplay-cli/commit/c71c4acbd98f3a5916b915c791d136fd8141fa7f))
* **deps:** update module github.com/spf13/cobra to v1.10.2 ([#98](https://github.com/nentgroup/viaplay-cli/issues/98)) ([6e2ae41](https://github.com/nentgroup/viaplay-cli/commit/6e2ae41df596ac68cf54246a4578560ad974ef40))
* **deps:** update module github.com/spf13/viper to v1.21.0 ([#69](https://github.com/nentgroup/viaplay-cli/issues/69)) ([5ee8dd2](https://github.com/nentgroup/viaplay-cli/commit/5ee8dd287eb331a569637912bb5807c6521be121))
* **deps:** update module golang.org/x/crypto to v0.42.0 ([#71](https://github.com/nentgroup/viaplay-cli/issues/71)) ([7444411](https://github.com/nentgroup/viaplay-cli/commit/7444411abaf9bd5ed4ef726b05af9a4c9efcf767))
* **deps:** update module golang.org/x/crypto to v0.43.0 ([#78](https://github.com/nentgroup/viaplay-cli/issues/78)) ([5077f5a](https://github.com/nentgroup/viaplay-cli/commit/5077f5a5387fe88a2064480753bea7ce5bbd88ab))
* **deps:** update module golang.org/x/crypto to v0.44.0 ([#90](https://github.com/nentgroup/viaplay-cli/issues/90)) ([8947f52](https://github.com/nentgroup/viaplay-cli/commit/8947f526e0603e872288be4629126cf774168f04))
* **deps:** update module golang.org/x/crypto to v0.45.0 [security] ([#96](https://github.com/nentgroup/viaplay-cli/issues/96)) ([f4f52ce](https://github.com/nentgroup/viaplay-cli/commit/f4f52ce66bd3c20ab6c6f248b04a39635bf65a39))
* **deps:** update module golang.org/x/crypto to v0.46.0 ([#103](https://github.com/nentgroup/viaplay-cli/issues/103)) ([699006e](https://github.com/nentgroup/viaplay-cli/commit/699006e46b621c322d1435863b496af99d793b72))
* **deps:** update module golang.org/x/oauth2 to v0.31.0 ([#66](https://github.com/nentgroup/viaplay-cli/issues/66)) ([05c6d8d](https://github.com/nentgroup/viaplay-cli/commit/05c6d8deea1f045cd6da738c3e9e34265f5b11b3))
* **deps:** update module golang.org/x/oauth2 to v0.32.0 ([#76](https://github.com/nentgroup/viaplay-cli/issues/76)) ([f17e20a](https://github.com/nentgroup/viaplay-cli/commit/f17e20a84c49470b755a144d3a8c4fc0d04a0cda))
* **deps:** update module golang.org/x/oauth2 to v0.33.0 ([#86](https://github.com/nentgroup/viaplay-cli/issues/86)) ([d4a8946](https://github.com/nentgroup/viaplay-cli/commit/d4a89462a4876ed9508f87bcf0ae6e56c065b772))
* **deps:** update module golang.org/x/oauth2 to v0.34.0 ([#102](https://github.com/nentgroup/viaplay-cli/issues/102)) ([419afd7](https://github.com/nentgroup/viaplay-cli/commit/419afd7af6f5a8492bd1be4633a09469043933ad))
* **deps:** update module golang.org/x/term to v0.35.0 ([#67](https://github.com/nentgroup/viaplay-cli/issues/67)) ([bf58c83](https://github.com/nentgroup/viaplay-cli/commit/bf58c839df6300a85d84f5414dc197efb842ef06))

## [2.3.2](https://github.com/nentgroup/viaplay-cli/compare/v2.3.1...v2.3.2) (2025-09-04)


### Dependencies

* **deps:** update actions/labeler action to v6 ([#61](https://github.com/nentgroup/viaplay-cli/issues/61)) ([90b6dd9](https://github.com/nentgroup/viaplay-cli/commit/90b6dd9effc325dda6bdc65a8c6cd64974319168))

## [2.3.1](https://github.com/nentgroup/viaplay-cli/compare/v2.3.0...v2.3.1) (2025-09-04)


### Bug Fixes

* handle support for .env files correctly ([#62](https://github.com/nentgroup/viaplay-cli/issues/62)) ([d48fc7f](https://github.com/nentgroup/viaplay-cli/commit/d48fc7f923f8145362f9e36fb512be38814ae017))


### Dependencies

* **deps:** update actions/setup-go action to v6 ([#59](https://github.com/nentgroup/viaplay-cli/issues/59)) ([e2afc23](https://github.com/nentgroup/viaplay-cli/commit/e2afc232b99b5faad4a08b0b378c5156c3a8fee7))

## [2.3.0](https://github.com/nentgroup/viaplay-cli/compare/v2.2.0...v2.3.0) (2025-09-03)


### Features

* add json-flag to template test command ([#57](https://github.com/nentgroup/viaplay-cli/issues/57)) ([3171313](https://github.com/nentgroup/viaplay-cli/commit/31713137bfb364d62a68fb36fec53e1713efad17))

## [2.2.0](https://github.com/nentgroup/viaplay-cli/compare/v2.1.1...v2.2.0) (2025-09-03)


### Features

* add template test command ([#56](https://github.com/nentgroup/viaplay-cli/issues/56)) ([5252dc0](https://github.com/nentgroup/viaplay-cli/commit/5252dc0285e434b550f9d9788f9f2d9317f84dc7))


### Dependencies

* **deps:** update dependency font-awesome to v7.0.1 ([#55](https://github.com/nentgroup/viaplay-cli/issues/55)) ([4f421c6](https://github.com/nentgroup/viaplay-cli/commit/4f421c6a0df9f241bae9b16f6cda20f727a6cb00))
* **deps:** update module github.com/spf13/cobra to v1.10.1 ([#53](https://github.com/nentgroup/viaplay-cli/issues/53)) ([6ca8eda](https://github.com/nentgroup/viaplay-cli/commit/6ca8edad2156ab7cf2c31d3121810fc464ab4871))

## [2.1.1](https://github.com/nentgroup/viaplay-cli/compare/v2.1.0...v2.1.1) (2025-08-25)


### Documentation

* remove outdated references and apply minor corrections ([#51](https://github.com/nentgroup/viaplay-cli/issues/51)) ([4d303ff](https://github.com/nentgroup/viaplay-cli/commit/4d303ffa4f52d6ef785d2b957d358bfe4c441b1c))

## [2.1.0](https://github.com/nentgroup/viaplay-cli/compare/v2.0.5...v2.1.0) (2025-08-25)


### Features

* streamline config handling and improve CLI output ([#49](https://github.com/nentgroup/viaplay-cli/issues/49)) ([461ba48](https://github.com/nentgroup/viaplay-cli/commit/461ba48e5414a5bd4e6dc0d06eff37bfa960deff))

## [2.0.5](https://github.com/nentgroup/viaplay-cli/compare/v2.0.4...v2.0.5) (2025-08-19)


### Documentation

* update quickstart ([#47](https://github.com/nentgroup/viaplay-cli/issues/47)) ([31b4af6](https://github.com/nentgroup/viaplay-cli/commit/31b4af6be76b6da592d06994f297ebd76dbf5df0))

## [2.0.4](https://github.com/nentgroup/viaplay-cli/compare/v2.0.3...v2.0.4) (2025-08-19)


### Bug Fixes

* skip personal configs if they don't exists ([#45](https://github.com/nentgroup/viaplay-cli/issues/45)) ([d0750e1](https://github.com/nentgroup/viaplay-cli/commit/d0750e11cce89787562d4630c0c5226a8a3d67f2))

## [2.0.3](https://github.com/nentgroup/viaplay-cli/compare/v2.0.2...v2.0.3) (2025-08-19)


### Documentation

* remove deprecated references ([#43](https://github.com/nentgroup/viaplay-cli/issues/43)) ([18b2fbb](https://github.com/nentgroup/viaplay-cli/commit/18b2fbb51c772ee5b9c145eb30853dab0d4148d9))

## [2.0.2](https://github.com/nentgroup/viaplay-cli/compare/v2.0.1...v2.0.2) (2025-08-19)


### Bug Fixes

* restore banner ([b974e1b](https://github.com/nentgroup/viaplay-cli/commit/b974e1bb41b3f48ad799719bc2adfbd58ca29a47))

## [2.0.1](https://github.com/nentgroup/viaplay-cli/compare/v2.0.0...v2.0.1) (2025-08-19)


### Documentation

* update documentation ([5f03a38](https://github.com/nentgroup/viaplay-cli/commit/5f03a38bccbd37243c3addb22a1c51333265991d))

## [2.0.0](https://github.com/nentgroup/viaplay-cli/compare/v1.3.0...v2.0.0) (2025-08-19)


### ⚠ BREAKING CHANGES

* **config:** refactor configs ([#39](https://github.com/nentgroup/viaplay-cli/issues/39))

### Code Refactoring

* **config:** refactor configs ([#39](https://github.com/nentgroup/viaplay-cli/issues/39)) ([591acbb](https://github.com/nentgroup/viaplay-cli/commit/591acbbfaa4c15e8d5fa511f06ea17bc736c81c3))

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
