const semver = require("semver")
const fs = require("fs/promises")
const path = require("path")

const pkg = require("../package.json")

async function updatePackageVersion() {
  try {
    const gitRef = process.env.GIT_REFNAME
    const gitTag = semver.valid(gitRef);
    if (gitTag && semver.compare(gitTag, pkg.version) > 0) {
      pkg.version = gitTag;
      console.log('Updating JS SDK version to', gitTag)
      await fs.writeFile(path.join(process.cwd(), 'package.json'), JSON.stringify(pkg, null, 2))
    }
  } catch (err) {
    console.error("Update Package Version", err)
  }
}

updatePackageVersion()