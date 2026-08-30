# Changelog

## [4.0.0](https://github.com/nblair2/go-dnp3/compare/v3.0.0...v4.0.0) (2026-08-30)


### ⚠ BREAKING CHANGES

* decode event flags, status, and time as first-class fields ([#70](https://github.com/nblair2/go-dnp3/issues/70))

### Features

* **assemble:** add AssemblePayload for TCP payload ([#61](https://github.com/nblair2/go-dnp3/issues/61)) ([05dc362](https://github.com/nblair2/go-dnp3/commit/05dc362fd15c3d5e5fdf32a797d232e752ff7c1c))
* decode event flags, status, and time as first-class fields ([#70](https://github.com/nblair2/go-dnp3/issues/70)) ([eb9d8cd](https://github.com/nblair2/go-dnp3/commit/eb9d8cd97978af9fa302544fd5a715d5ef34d648)), closes [#46](https://github.com/nblair2/go-dnp3/issues/46)


### Bug Fixes

* accept read-all variation-0 requests for all groups ([#68](https://github.com/nblair2/go-dnp3/issues/68)) ([f2c22f8](https://github.com/nblair2/go-dnp3/commit/f2c22f84dd6b7ff100bbf83f676f2190c9a2d02e)), closes [#45](https://github.com/nblair2/go-dnp3/issues/45)
* bound updateIndexes iteration to prevent uint32 wrap ([#66](https://github.com/nblair2/go-dnp3/issues/66)) ([0fa43e1](https://github.com/nblair2/go-dnp3/commit/0fa43e1bca30680358012ed5f3b83e4834929c32)), closes [#30](https://github.com/nblair2/go-dnp3/issues/30)
* correct byte index and masks in newPoints2Bits ([#67](https://github.com/nblair2/go-dnp3/issues/67)) ([19cfffc](https://github.com/nblair2/go-dnp3/commit/19cfffc94161941e66edb895749c1df6801bf2ec)), closes [#31](https://github.com/nblair2/go-dnp3/issues/31)
* decode g50v2 absolute time before interval ([#72](https://github.com/nblair2/go-dnp3/issues/72)) ([785ffc5](https://github.com/nblair2/go-dnp3/commit/785ffc58d5c9e99ef80674b048e07ed2260bf9f0)), closes [#71](https://github.com/nblair2/go-dnp3/issues/71)
* guard ApplicationRequest.DecodeFromBytes against short input ([#65](https://github.com/nblair2/go-dnp3/issues/65)) ([ce85a7a](https://github.com/nblair2/go-dnp3/commit/ce85a7acc0471a92f70c8b71a55915b780c493d0)), closes [#33](https://github.com/nblair2/go-dnp3/issues/33)
* honor index prefix in newPointsBitFlags ([#73](https://github.com/nblair2/go-dnp3/issues/73)) ([27c400e](https://github.com/nblair2/go-dnp3/commit/27c400ef0500d2b5329f3528e221e4395722e6a1)), closes [#32](https://github.com/nblair2/go-dnp3/issues/32)

## [3.0.0](https://github.com/nblair2/go-dnp3/compare/v2.1.0...v3.0.0) (2026-08-23)


### ⚠ BREAKING CHANGES

* migrate to gopacket/gopacket fork ([#57](https://github.com/nblair2/go-dnp3/issues/57))

### Miscellaneous Chores

* migrate to gopacket/gopacket fork ([#57](https://github.com/nblair2/go-dnp3/issues/57)) ([c622ffe](https://github.com/nblair2/go-dnp3/commit/c622ffe6b3341fe0ba3d2b93e87b60a29c3931ab)), closes [#52](https://github.com/nblair2/go-dnp3/issues/52)

## [2.1.0](https://github.com/nblair2/go-dnp3/compare/v2.0.2...v2.1.0) (2026-08-23)


### Features

* transport-layer reassembly / session API ([#58](https://github.com/nblair2/go-dnp3/issues/58)) ([7f19d0a](https://github.com/nblair2/go-dnp3/commit/7f19d0a79b067f82e3f22cc16afca58030b0dcad)), closes [#50](https://github.com/nblair2/go-dnp3/issues/50)

## [2.0.2](https://github.com/nblair2/go-dnp3/compare/v2.0.1...v2.0.2) (2026-08-23)


### Bug Fixes

* guard decoders against short input ([3b91880](https://github.com/nblair2/go-dnp3/commit/3b918800ea0767e0454973a98523e7fa86e1e202))
* objectType data race and sequence off-by-one ([2ca6942](https://github.com/nblair2/go-dnp3/commit/2ca694240012f58f23858f979b46d30a5fb1543c))
* wire-format decode and round-trip bugs ([fc3fc87](https://github.com/nblair2/go-dnp3/commit/fc3fc873dc39f4533280a445ab4e078390ce9563))

## [2.0.1](https://github.com/nblair2/go-dnp3/compare/v2.0.0...v2.0.1) (2026-05-15)


### Bug Fixes

* add v2 suffix in go.mod ([#26](https://github.com/nblair2/go-dnp3/issues/26)) ([f15817d](https://github.com/nblair2/go-dnp3/commit/f15817daa951d117811782821bd7e45aa09b6385))

## [2.0.0](https://github.com/nblair2/go-dnp3/compare/v1.2.0...v2.0.0) (2026-05-13)


### ⚠ BREAKING CHANGES

* FromBytes / ToBytes --> DecodeFromBytes / SerializeTo

### Features

* gopacket compliance ([#25](https://github.com/nblair2/go-dnp3/issues/25)) ([59d4f97](https://github.com/nblair2/go-dnp3/commit/59d4f97f17f2ab782ce9105ba8041c6fbd509182))
* multi-frame parsing, extra accessors ([#22](https://github.com/nblair2/go-dnp3/issues/22)) ([ec3d6bd](https://github.com/nblair2/go-dnp3/commit/ec3d6bd28e7816b4cc78132edf0733f0f2be6c96))

## [1.2.0](https://github.com/nblair2/go-dnp3/compare/v1.1.0...v1.2.0) (2026-05-09)


### Features

* add New* and New*FromBytes constructor pairs ([#19](https://github.com/nblair2/go-dnp3/issues/19)) ([de5dd2e](https://github.com/nblair2/go-dnp3/commit/de5dd2e6502c507969640392af72e4c7b9ea6656))

## [1.1.0](https://github.com/nblair2/go-dnp3/compare/v1.0.1...v1.1.0) (2026-04-05)


### Features

* point accessors ([#15](https://github.com/nblair2/go-dnp3/issues/15)) ([7ab8e2d](https://github.com/nblair2/go-dnp3/commit/7ab8e2dd7e67a924bba824668482ac3bcd5bf9ef))

## [1.0.1](https://github.com/nblair2/go-dnp3/compare/v1.0.0...v1.0.1) (2025-12-31)


### Bug Fixes

* PointNBytesFlag, DataLinkControlFucntionCodes ([#10](https://github.com/nblair2/go-dnp3/issues/10)) ([ae7d1da](https://github.com/nblair2/go-dnp3/commit/ae7d1dad8e8a17b84f7fae8a804faba5d77f68ff))

## [1.0.0](https://github.com/nblair2/go-dnp3/compare/v0.1.2...v1.0.0) (2025-11-30)


### ⚠ BREAKING CHANGES

* Some FromBytes / ToBytes function signatures include errors where they previously did not

### Code Refactoring

* better errors, organization, names ([#8](https://github.com/nblair2/go-dnp3/issues/8)) ([9c34c76](https://github.com/nblair2/go-dnp3/commit/9c34c76f63dea896b7dfa9d0fab60649a1471581))

## [0.1.2](https://github.com/nblair2/go-dnp3/compare/v0.1.1...v0.1.2) (2025-10-05)


### Bug Fixes

* correct module path ([#4](https://github.com/nblair2/go-dnp3/issues/4)) ([b95f244](https://github.com/nblair2/go-dnp3/commit/b95f244697bfcef1df6c1e6dc6a154f069567fb1))

## [0.1.1](https://github.com/nblair2/go-dnp3/compare/v0.1.0...v0.1.1) (2025-10-05)


### Bug Fixes

* **CI:** ci action on PR accept -&gt; release action ([#2](https://github.com/nblair2/go-dnp3/issues/2)) ([3c09689](https://github.com/nblair2/go-dnp3/commit/3c0968984307b88081afb7ba98064eb558851188))
