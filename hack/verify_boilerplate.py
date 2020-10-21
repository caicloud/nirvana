#!/usr/bin/env python

# Copyright 2017 Caicloud Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Verifies that all source files contain the necessary copyright boilerplate
# snippet.

from __future__ import print_function

import argparse
import glob
import os
import re
import sys

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "filenames", help="list of files to check, all files if unspecified", nargs='*')

    rootdir = os.path.dirname(__file__) + "/../"
    rootdir = os.path.abspath(rootdir)
    parser.add_argument("--rootdir", default=rootdir, help="root directory to examine")

    default_boilerplate_dir = os.path.join(rootdir, "hack/boilerplate")
    parser.add_argument("--boilerplate-dir", default=default_boilerplate_dir)
    return parser.parse_args()


def get_refs():
    refs = {}

    for path in glob.glob(os.path.join(ARGS.boilerplate_dir, "boilerplate.*.txt")):
        extension = os.path.basename(path).split(".")[1]

        ref_file = open(path, 'r')
        ref = ref_file.read().splitlines()
        ref_file.close()
        refs[extension] = ref

    return refs


def file_passes(filename, refs, regexs):  # pylint: disable=too-many-locals
    try:
        with open(filename, 'r') as fp:
            data = fp.read()
    except IOError:
        return False

    basename = os.path.basename(filename)
    extension = file_extension(filename)
    if extension != "":
        ref = refs[extension]
    else:
        ref = refs[basename]

    # remove build tags from the top of Go files
    if extension == "go":
        con = regexs["go_build_constraints"]
        (data, found) = con.subn("", data, 1)

    # remove shebang from the top of shell files
    if extension == "sh" or extension == "py":
        she = regexs["shebang"]
        (data, found) = she.subn("", data, 1)

    data = data.splitlines()

    # if our test file is smaller than the reference it surely fails!
    if len(ref) > len(data):
        return False

    # trim our file to the same number of lines as the reference file
    data = data[:len(ref)]

    year = regexs["year"]
    for datum in data:
        if year.search(datum):
            return False

    # Replace all occurrences of the regex "2017|2018" with "YEAR"
    when = regexs["date"]
    for idx, datum in enumerate(data):
        (data[idx], found) = when.subn('YEAR', datum)
        if found != 0:
            break

    # if we don't match the reference at this point, fail
    if ref != data:
        return False

    return True

def file_extension(filename):
    return os.path.splitext(filename)[1].split(".")[-1].lower()

SKIPPED_DIRS = ['Godeps', 'third_party', '_gopath', '_output', '.git', 'vendor', '__init__.py']

def normalize_files(files):
    newfiles = []
    for pathname in files:
        if any(x in pathname for x in SKIPPED_DIRS):
            continue
        newfiles.append(pathname)
    for idx, pathname in enumerate(newfiles):
        if not os.path.isabs(pathname):
            newfiles[idx] = os.path.join(ARGS.rootdir, pathname)
    return newfiles


def get_files(extensions):
    files = []
    if ARGS.filenames:
        files = ARGS.filenames
    else:
        for root, dirs, walkfiles in os.walk(ARGS.rootdir):
            # don't visit certain dirs. This is just a performance improvement
            # as we would prune these later in normalize_files(). But doing it
            # cuts down the amount of filesystem walking we do and cuts down
            # the size of the file list
            for dpath in SKIPPED_DIRS:
                if dpath in dirs:
                    dirs.remove(dpath)

            for name in walkfiles:
                pathname = os.path.join(root, name)
                files.append(pathname)

    files = normalize_files(files)
    outfiles = []
    for pathname in files:
        basename = os.path.basename(pathname)
        extension = file_extension(pathname)
        if extension in extensions or basename in extensions:
            outfiles.append(pathname)
    return outfiles


def get_regexs():
    regexs = {}
    # Search for "YEAR" which exists in the boilerplate, but shouldn't in the real thing
    regexs["year"] = re.compile('YEAR')
    # dates can be 2017 or 2018, company holder names can be anything
    regexs["date"] = re.compile('(2017|2018)')
    # strip // +build \n\n build constraints
    regexs["go_build_constraints"] = re.compile(r"^(// \+build.*\n)+\n", re.MULTILINE)
    # strip #!.* from shell/python scripts
    regexs["shebang"] = re.compile(r"^(#!.*\n)\n*", re.MULTILINE)
    return regexs


def main():
    regexs = get_regexs()
    refs = get_refs()
    filenames = get_files(refs.keys())
    nonconforming_files = []
    for filename in filenames:
        if not file_passes(filename, refs, regexs):
            nonconforming_files.append(filename)

    if nonconforming_files:
        print('%d files have incorrect boilerplate headers:' %
              len(nonconforming_files))
        for filename in sorted(nonconforming_files):
            print(filename)
        sys.exit(1)


if __name__ == "__main__":
    ARGS = get_args()
    main()
