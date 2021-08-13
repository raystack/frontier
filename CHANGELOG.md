# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [0.1.5-rc.0](https://github.com/odpf/shield/compare/v0.1.4...v0.1.5-rc.0) (2021-08-13)


### Features

* add redis caching for group list api ([d53b7d6](https://github.com/odpf/shield/commit/d53b7d657cf432831457e6c37b31c037b9e23818))


### Bug Fixes

* logs for http request ([b2a9fc6](https://github.com/odpf/shield/commit/b2a9fc6bf94ec30cc865c24be3e1ba86f6f338c0))

### [0.1.4](https://github.com/odpf/shield/compare/v0.1.3...v0.1.4) (2021-08-09)


### Bug Fixes

* check for more than 0 role tags for filter ([4ae69b1](https://github.com/odpf/shield/commit/4ae69b15eecfedf86cdd64f03f8d2de0b1c08b43))

### [0.1.3](https://github.com/odpf/shield/compare/v0.1.2...v0.1.3) (2021-08-06)


### Bug Fixes

* typo in conf file destination postbuild ([89c455c](https://github.com/odpf/shield/commit/89c455c0618f844c2431b2a9fafbc0f5a85b0d24))

### [0.1.2](https://github.com/odpf/shield/compare/v0.1.1...v0.1.2) (2021-08-06)


### Bug Fixes

* build folder structure ([ea3f9cc](https://github.com/odpf/shield/commit/ea3f9cce164a71a0b4866b6de818b89cc9af9be0))
* policy filter in get all user api ([#16](https://github.com/odpf/shield/issues/16)) ([37e6911](https://github.com/odpf/shield/commit/37e69115942c2c0284aa5fec5559a3edafc9c6f2))

### [0.1.1](https://github.com/odpf/shield/compare/v0.1.0...v0.1.1) (2021-07-12)


### Features

* add tags filtering for roles ([#10](https://github.com/odpf/shield/issues/10)) ([85ca9fa](https://github.com/odpf/shield/commit/85ca9fa43c7b1fa9ff2737dcd575daad2c02b837))

## 0.1.0 (2021-05-24)


### Features

* add default role for when users are added to group ([8a63795](https://github.com/odpf/shield/commit/8a63795d0f5ae27ea6b89133fe3ac30389a9d910))
* add memberCount to get group by id API ([88a859d](https://github.com/odpf/shield/commit/88a859d08fae4ec443876442bd0c34090d7d6e7d))
* add newrelic for monitoring ([0a8e6b7](https://github.com/odpf/shield/commit/0a8e6b7d371e583394936da432e59394db6bf9c2))
* add resource attributes mapping create/delete API ([f6569a5](https://github.com/odpf/shield/commit/f6569a578d393473a98c0bc0c97aabef3156c1b0))
* add role response to update role API ([2832e26](https://github.com/odpf/shield/commit/2832e264814b2bb63810239ca275c313e491d3ac))
* add test github action ([b100853](https://github.com/odpf/shield/commit/b100853e8f4d54a234be6b3e0fdc953ac4209752))
* add tests for both loadFiltered and loadAll policies ([7642680](https://github.com/odpf/shield/commit/76426809e74a5b17c29cb48dc3cdf7291436d57b))
* add unknown for response validation ([a1f1390](https://github.com/odpf/shield/commit/a1f13904108602fecfeadc55f8424eeedef4f8a2))
* add user-id in reverse proxy headers ([10839ca](https://github.com/odpf/shield/commit/10839cafb5dcdba785bf3328acb250003c1ff0bb))
* add username and groupname columns ([5f7ffb0](https://github.com/odpf/shield/commit/5f7ffb0d0cfccf1ea13ea70b0cb5ee91eab7846a))
* add validation for username and groupname ([16e092b](https://github.com/odpf/shield/commit/16e092bd17a0e159db28273e62849d96672eef6c))
* allow unknown query parameters ([2c7e357](https://github.com/odpf/shield/commit/2c7e35772cb4972fdee70a39a650a716da807fb0))
* delete generic model subscriber, add group subscribe for group related event ([0278661](https://github.com/odpf/shield/commit/0278661476c82eaa0cb69afa13174fca41fe0372))
* fix broken test cases ([d89b571](https://github.com/odpf/shield/commit/d89b571764981a436f06ec8296f13d3e9672d3d2))
* further optimize policy subset loading ([d4087b6](https://github.com/odpf/shield/commit/d4087b656b9c15866a2c422d0b6b56da30b39050))
* get activities based on group , test cases for activity handler and resources ([e6d6984](https://github.com/odpf/shield/commit/e6d6984dce6854790e3cdb9ccb241de3016e5356))
* get all activity log or based on group name ([a0f4108](https://github.com/odpf/shield/commit/a0f41082b9952b85062501fdda2d263beeefbf25))
* get list of groups based on attribute filters ([15fa751](https://github.com/odpf/shield/commit/15fa7510ed2f3f6d49c26ff8450735c37b043460))
* get list of groups of a user ([9852001](https://github.com/odpf/shield/commit/9852001de1d4afc6e16d0d571980faed6d067382))
* get list of users ([2960125](https://github.com/odpf/shield/commit/29601255c94eab7a32c09467ca2c0dcd9d665f8f))
* get users of a group with policies ([0a318e4](https://github.com/odpf/shield/commit/0a318e4a6ca2908e4f1fd3c3e5c81a4982fd7bef))
* get users of a group with policies with policy filters ([b0568c3](https://github.com/odpf/shield/commit/b0568c3261d685d24939332b8aada1171f0b472d))
* group proxy related api in swagger ([c21bde4](https://github.com/odpf/shield/commit/c21bde4b6b4df3730886c667e684724ffd7c713a))
* implement ability to load subset of a policy ([7d0c456](https://github.com/odpf/shield/commit/7d0c45671a0e162e0e5e35f8af3e738a68c4f278))
* implement check-access API ([40e6a42](https://github.com/odpf/shield/commit/40e6a427cfcee12eafc30cd477e11c883ef06f98))
* implement create/update role along with action API ([10cbeed](https://github.com/odpf/shield/commit/10cbeed5e3e3ff2a09aee383c9c7d726a9df8e75))
* implement role creation along with action mapping API ([77b96e3](https://github.com/odpf/shield/commit/77b96e3f8a0f372d9d61fc203d74a395fc8125a5))
* Integrate Joi validation in profile + metadata schema API ([4055799](https://github.com/odpf/shield/commit/40557997d50ac9ae4383ea6c3da817ee9134cba0))
* introduce batch enforce json policies ([935f4b8](https://github.com/odpf/shield/commit/935f4b87778ae5e321c671f0336148ee6d03d274))
* introduce new options for proxy ([5c27f3c](https://github.com/odpf/shield/commit/5c27f3c57ea4a57ff68e0e2a60b6ec4efda141e7))
* log activity for team create and edit ([4978f75](https://github.com/odpf/shield/commit/4978f754298e40f27aa2433356f90436bb474695))
* log activity without calling db for policies ([280a3a0](https://github.com/odpf/shield/commit/280a3a0aea90a721ea54188b69a869cbeb83ab18))
* log roles in activity ([0e08488](https://github.com/odpf/shield/commit/0e084880ae05e7db1580bf9842f31b5be225fe6c))
* merge resource attributes along with response using hooks ([10a81d9](https://github.com/odpf/shield/commit/10a81d93079eae467817ced1852471e0bf073c70))
* pass logged in user into group user api ([c0e7142](https://github.com/odpf/shield/commit/c0e714237c1a4c1ac6f8cf62dc1b412d710a7186))
* readme file for shield ([33814e2](https://github.com/odpf/shield/commit/33814e2a78b5f58e2fbfd16b496e8a9c23cc2207))
* remove policies log ([ca46660](https://github.com/odpf/shield/commit/ca46660f47b5062554288de2d962470e04d21028))
* remove self from a group api ([34aa993](https://github.com/odpf/shield/commit/34aa99361d3f68263b7d60d1b79ca1ac0cf7f2f4))
* remove unused methods ([383a0ec](https://github.com/odpf/shield/commit/383a0ecbb63aea958215d1ef28dc5925f5f9a091))
* resolve merge conflicts ([9a06f56](https://github.com/odpf/shield/commit/9a06f569ec7470a728523c3beea18f4c435a99c5))
* return downstream error response in proxy ([b1eb689](https://github.com/odpf/shield/commit/b1eb68959ce9b047ae0a757c486a18a452c738bd))
* swagger documentation for shield apis ([b7120e6](https://github.com/odpf/shield/commit/b7120e6045c63b968c795e169c326a6aa4a3b2b6))
