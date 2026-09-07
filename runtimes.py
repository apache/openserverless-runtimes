#!/usr/bin/env python3
"""Generate runtimes.json by walking the runtime/ subtree.
/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

  Each runtime/<language>/<version> directory becomes one kind: the language
  directory gives the kind family, the version subdir gives both the kind
  version and the image tag, with -<tag> appended. A leading `v` is optional
  and stripped from the kind, so runtime/python/v3.12 yields python:3.12 and
  runtime/python/sys yields python:sys. A merge.json in a version directory is
  merged into that entry, which is how defaults, stemCells and requireMain are
  declared.

  Example:
     ./runtime.py all_26i06r36-snapshot
     ./runtime.py all_26i06r36-snapshot -o runtimes.json
"""

import argparse
import json
import os
import sys

DESCRIPTION = [
    "This file describes the different languages (aka. managed action runtimes) supported by the system",
    "as well as blackbox images that support the runtime specification.",
    "Only actions with runtime families / kinds defined here can be created / read / updated / deleted / invoked.",
    "Define a list of runtime families (example: 'nodejs') with at least one kind per family (example: 'nodejs:14').",
    "Each runtime family needs a default kind (default: true).",
    "When removing or renaming runtime families or runtime kinds from this file, preexisting actions",
    "with the affected kinds can no longer be read / updated / deleted / invoked. In order to remove or rename",
    "runtime families or runtime kinds, mark all affected runtime kinds as deprecated (deprecated: true) and",
    "perform a manual migration of all affected actions.",
    "",
    "This file is meant to list all stable runtimes supported by the Apache Openwhisk community.",
]

ATTACHED = {"attachmentName": "codefile", "attachmentType": "text/plain"}


def read_env(path):
    """Parse a dotenv file into a dict; missing file yields an empty environment."""
    env = {}
    if not os.path.isfile(path):
        return env
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, value = line.split("=", 1)
            env[key.strip()] = value.strip().strip('"').strip("'")
    return env


def strip_prefix(tag):
    """Drop the leading `xxx_` build selector, e.g. all_26i06r36-snapshot -> 26i06r36-snapshot."""
    return tag.split("_", 1)[1] if "_" in tag else tag


def version_key(version):
    """Sort v3.11/v1.27/v8 numerically, newest first, so ordering is stable.

    Named variants such as `sys` carry no version number; they sort after the
    numbered ones. Every component is compared as a (rank, value) pair so a
    numbered part never has to compare against a textual one.
    """
    numbered = version.startswith("v")
    parts = [
        (0, p.zfill(8)) if p.isdigit() else (1, p)
        for p in version.lstrip("v").split(".")
    ]
    return [(0,) if numbered else (1,), parts]


def kind_entry(language, version, tag, prefix):
    """Build one kind from a runtime/<language>/<version> directory."""
    entry = {
        "kind": "%s:%s" % (language, version.lstrip("v")),
        "default": False,
        "image": {
            "prefix": prefix,
            "name": "openserverless-runtime-%s" % language,
            "tag": "%s-%s" % (version, tag),
        },
        "deprecated": False,
        "attached": dict(ATTACHED),
    }
    return entry


def walk(runtime_dir, tag, prefix):
    """Walk runtime/<language>/<version>, yielding (language, [kinds])."""
    for language in sorted(os.listdir(runtime_dir)):
        lang_dir = os.path.join(runtime_dir, language)
        if not os.path.isdir(lang_dir):
            continue

        versions = sorted(
            (
                d
                for d in os.listdir(lang_dir)
                if not d.startswith(".")
                and os.path.isdir(os.path.join(lang_dir, d))
            ),
            key=version_key,
            reverse=True,
        )
        if not versions:
            continue

        kinds = []
        for version in versions:
            entry = kind_entry(language, version, tag, prefix)
            merge_file = os.path.join(lang_dir, version, "merge.json")
            if os.path.isfile(merge_file):
                with open(merge_file) as f:
                    try:
                        entry.update(json.load(f))
                    except ValueError as err:
                        sys.exit("ERROR: cannot parse %s: %s" % (merge_file, err))
            kinds.append(entry)
        yield language, kinds


def build(runtime_dir, tag, prefix):
    runtimes = {}
    for language, kinds in walk(runtime_dir, tag, prefix):
        runtimes[language] = kinds
    return {"description": DESCRIPTION, "runtimes": runtimes, "blackboxes": []}


def main():
    here = os.path.dirname(os.path.abspath(__file__))
    parser = argparse.ArgumentParser(
        description="generate runtimes.json by walking the runtime/ subtree"
    )
    parser.add_argument(
        "tag", help="build tag, e.g. all_26i06r36-snapshot (the xxx_ prefix is removed)"
    )
    parser.add_argument(
        "--env",
        default=os.path.join(here, ".env"),
        help="dotenv file providing DOCKER_HUB_REGISTRY and DOCKERHUB_USER",
    )
    parser.add_argument(
        "--runtime-dir",
        default=os.path.join(here, "runtime"),
        help="path to the runtime/ tree (default: alongside this script)",
    )
    parser.add_argument("-o", "--output", help="write to this file instead of stdout")
    args = parser.parse_args()

    # .env provides the defaults; a real environment variable still wins.
    env = read_env(args.env)
    registry = os.environ.get(
        "DOCKER_HUB_REGISTRY", env.get("DOCKER_HUB_REGISTRY", "docker.io")
    )
    user = os.environ.get("DOCKERHUB_USER", env.get("DOCKERHUB_USER", "apache"))
    prefix = "%s/%s" % (registry, user)

    result = build(args.runtime_dir, strip_prefix(args.tag), prefix)
    text = json.dumps(result, indent=4) + "\n"

    if args.output:
        with open(args.output, "w") as f:
            f.write(text)
        print("Generated %s" % args.output, file=sys.stderr)
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main()
