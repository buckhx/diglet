#!usr/bin/env python

'''
This script has no dependencies besides a connection to github in order to run.
Therefore it can be copied from the repo and ran on any unix system
By default it will install to /usr/local/bin/diglet.
'''
import json
import os
import os.path
import platform
import sys
import urllib

APP = "diglet"
DEFAULT_INSTALL = "/usr/local/bin/"

# Change this if you don't have root to somewhere else on your path
if len(sys.argv) > 1:
    bindir = os.path.join(sys.argv[1], APP)
else:
    bindir = os.path.join(DEFAULT_INSTALL, APP)
print "Installing {0} into {1}...".format(APP, bindir)

kernel, _, _, _, arch, _ = platform.uname()
if arch == 'x86_64':
    arch = 'amd64'
elif arch == 'i386':
    arch = '386'
dist = "{0}_{1}".format(kernel.lower(), arch.lower())
artifact = "{0}_{1}".format(APP, dist)

#TODO remove static url
content = urllib.urlopen('https://api.github.com/repos/buckhx/diglet/releases/latest').read()
release = json.loads(content)
link = [asset['browser_download_url'] for asset in release['assets'] if asset['name'] == artifact][0]
print "Downloading binary from " + link
urllib.urlretrieve(link, bindir)
os.chmod(bindir, 0755)
print "Installed {0} at {1}".format(APP, bindir)
