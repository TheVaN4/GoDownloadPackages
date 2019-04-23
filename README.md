# Project state
The project is in the process of writing.
Development is conducted for learning and fun ;-)

# GoDownloadPackages
[![Build Status](https://travis-ci.org/thevan4/GoDownloadPackages.svg?branch=master)](https://travis-ci.org/thevan4/GoDownloadPackages) [![Go Report Card](https://goreportcard.com/badge/github.com/thevan4/GoDownloadPackages)](https://goreportcard.com/report/github.com/thevan4/GoDownloadPackages)

Uploading packages to ubuntu using the utility apt.
The program generate threads for cyclically and recursively creates a dependency map for all packages that may be required for installation. Dependencies of dependencies of dependencies... Blah blah)
No "sudo" required.

Example: 
./dl-deb -pc gnome-icon-theme (prints all packages) 
./dl-deb -pc gnome-icon-theme -dl yes (next to the binary creates a folder "./packs/" and downloads packages there) 
./dl-deb -pc gnome-icon-theme -dl yes -fl /somePath/ (downloads packages there) 
