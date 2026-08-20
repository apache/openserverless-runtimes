#!/usr/bin/env python3
"""Render runtimes.json.tpl and enrich it with runtime folder / endoflife.date metadata."""

import json
import os
import re
import urllib.error
import urllib.request

REPO_ROOT = os.path.dirname(os.path.abspath(__file__))
RUNTIME_DIR = os.path.join(REPO_ROOT, "runtime")
TEMPLATE_FILE = os.path.join(REPO_ROOT, "runtimes.json.tpl")
OUTPUT_FILE = os.path.join(REPO_ROOT, "runtimes.json")

LANGUAGES = ["go", "java", "nodejs", "python", "php"]

# endoflife.date product slug used by each language (java runs on IBM Semeru).
ENDOFLIFE_PRODUCTS = {
    "go": "go",
    "java": "ibm-semeru-runtime",
    "nodejs": "nodejs",
    "python": "python",
    "php": "php",
}

ENDOFLIFE_API = "https://endoflife.date/api/v1/products/{product}/"

ENV_VAR_PATTERN = re.compile(r"\$\{(\w+)\}|\$(\w+)")


def substitute_env(text):
    """
    Replace $VAR / ${VAR} placeholders with values from the environment,
    same behavior as envsubst: an unset variable becomes an empty string.
    """
    def replace(match):
        name = match.group(1) or match.group(2)
        return os.environ.get(name, "")

    return ENV_VAR_PATTERN.sub(replace, text)


def version_from_tag(tag_value):
    """
    Derive the runtime/<language>/<version> folder name from a resolved
    image tag (e.g. "v3.13-2608131945" -> "v3.13").
    """
    return tag_value.split("-", 1)[0] if tag_value else ""


def fetch_releases(product):
    """
    Fetch release cycles for a product from endoflife.date, keyed by cycle name.
    Returns an empty dict if the API call fails.
    """
    url = ENDOFLIFE_API.format(product=product)
    try:
        with urllib.request.urlopen(url, timeout=10) as response:
            data = json.load(response)
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
        print(f"Warning: could not fetch endoflife.date data for '{product}': {exc}")
        return {}

    return {release["name"]: release for release in data["result"]["releases"]}


def release_status(release):
    """
    Map an endoflife.date release cycle to "lts" or "current", following the same
    active-support/security-only distinction the endoflife.date UI uses.
    """
    if release.get("isLts"):
        return "lts"
    if release.get("isMaintained") and release.get("isEoas") is not True:
        return "current"
    return None


def has_files(path):
    """
    Check if folder has files inside it
    """
    for _, _, files in os.walk(path):
        if files:
            return True
    return False


def is_extendible(version_path):
    """
    Check if runtime folder has the extend script in bin/extend or extend.
    """
    return os.path.isfile(os.path.join(version_path, "bin", "extend")) or os.path.isfile(
        os.path.join(version_path, "extend")
    )


def scan_language(language, releases):
    versions = {}
    language_dir = os.path.join(RUNTIME_DIR, language)
    if not os.path.isdir(language_dir):
        return versions

    for entry in sorted(os.listdir(language_dir)):
        version_path = os.path.join(language_dir, entry)
        # Skip if the entry is not a directory
        if not os.path.isdir(version_path):
            continue
        # Skip if the folder is empty
        if not has_files(version_path):
            continue

        version_info = {}
        if is_extendible(version_path):
            version_info["extendible"] = True

        # Folder names are prefixed with "v" (e.g. "v3.13"), endoflife.date cycle
        # names are not (e.g. "3.13").
        release = releases.get(entry.lstrip("v"))
        if release:
            status = release_status(release)
            if status == "lts":
                version_info["lts"] = True
            elif status == "current":
                version_info["current"] = True

            if release.get("releaseDate"):
                version_info["releaseDate"] = release["releaseDate"]
            if release.get("latest", {}).get("name"):
                version_info["latest"] = release["latest"]["name"]

        versions[entry] = version_info

    return versions


def main():
    with open(TEMPLATE_FILE) as f:
        template_text = f.read()

    runtimes = json.loads(substitute_env(template_text))

    openserverless = {
        language: scan_language(language, fetch_releases(ENDOFLIFE_PRODUCTS[language]))
        for language in LANGUAGES
    }

    for language, kinds in runtimes.get("runtimes", {}).items():
        versions = openserverless.get(language, {})
        for kind in kinds:
            version = version_from_tag(kind.get("image", {}).get("tag", ""))
            if version in versions:
                kind["openserverless"] = versions[version]

    with open(OUTPUT_FILE, "w") as f:
        json.dump(runtimes, f, indent=4)
        f.write("\n")

    print(f"Written {OUTPUT_FILE}")


if __name__ == "__main__":
    main()
