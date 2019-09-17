# Development dependencies

To use a patched version of the `omci-sim` library:

```bash
make dep
cd vendor/github.com/opencord/
rm -rf omci-sim/
git clone https://gerrit.opencord.org/omci-sim
cd omci-sim
```

Once done, go to `gerrit.opencord.org` and locate the patch you want to get. Click on the download URL and copy the `Checkout` command.

It should look something like:

```
git fetch ssh://teone@gerrit.opencord.org:29418/omci-sim refs/changes/67/15067/1 && git checkout FETCH_HEAD
```

Then just execute that command in the `omci-sim` folder inside the vendored dependencies.