const semver = require('semver');

const pkg = require('../package.json');

function distTag(version) {
  const prerelease = semver.prerelease(version);
  if (!prerelease) return 'latest';
  const [preReleaseId] = prerelease;
  const isNamedTag =
    typeof preReleaseId === 'string' && !semver.validRange(preReleaseId);
  return isNamedTag ? preReleaseId : 'next';
}

console.log(distTag(pkg.version));
