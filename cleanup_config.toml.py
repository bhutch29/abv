#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Lexicographically sort the nicknames in config.toml

Dependencies:
  Python 3.4+
  The toml package (python -m pip install toml)

Example: Dry run in default location
  python cleanup_config.toml.py --dry-run

Example: Overwrite config.toml in working directory
  (Windows)   python cleanup_config.toml.py --dir=%CD%
  (Unix-like) python cleanup_config.toml.py --dir=$PWD
"""
import os.path as path
import sys

from argparse import ArgumentParser, RawTextHelpFormatter
from os.path import expanduser, realpath

import toml

CONFIG_TEMPLATE = """# Configuration file for ABV

# Barcodes should be strings, because leading zeros are not allowed
undoBarcode = "{undo}"
redoBarcode = "{redo}"

# Set location of config files. Do not use trailing slash. Defaults to ~/.abv
#configPath = ~/etc/abv

# Set the web root directory (directory containing the front page html and the static/ folder). Do not use trailing slash. Defaults to /srv/http
#webRoot = ~/.abv/www

{breweries}

{beers}
"""


def fullpath(*args):
    return realpath(path.join(*args))


def parsed_args():
    doc_lines = __doc__.strip().split('\n')
    parser = ArgumentParser(description=doc_lines[0],
                            epilog='\n'.join(doc_lines[1:]),
                            formatter_class=RawTextHelpFormatter)
    parser.add_argument('-d', '--dir',
                        action='store',
                        dest='dir',
                        default=fullpath(expanduser('~'), '.abv'),
                        help='directory containing config.toml (default: $HOME/.abv)')
    parser.add_argument('--dry-run',
                        dest='dry_run',
                        action='store_const',
                        const=True,
                        default=False,
                        help='prints the resulting config.toml file to stdout rather than overwriting')
    args = parser.parse_args()
    return args


def unpack_table(toml, name):
    # Format table title
    title = '[{name}]\n'.format(name=name)
    # Define a whitespace insertion method based on the longest key
    longest = max([len(key) for key in toml[name]])
    def align(str_):
        return ' '*(longest-len(str_)+1)
    # Lexicographically sort and whitespace-align table items
    table = sorted(toml[name].items(), key=lambda x: x[0])
    line = '"{key}"{ws}= "{val}"'
    items = '\n'.join([line.format(key=key, ws=align(key), val=val) for key, val in table])
    return title + items


def toml_from_file(filepath):
    try:
        parsed_toml = toml.load(filepath)
    except FileNotFoundError:
        sys.exit('File not found: {}'.format(filepath))
    return parsed_toml


def clean_config(filepath):
    parsed_toml = toml_from_file(filepath)
    # Unpack known config settings
    undo = parsed_toml['undoBarcode']
    redo = parsed_toml['redoBarcode']
    breweries = unpack_table(parsed_toml, 'breweryNicknames')
    beers = unpack_table(parsed_toml, 'beerNicknames')
    # Insert into config template
    text = CONFIG_TEMPLATE.format(undo=undo,
                                  redo=redo,
                                  breweries=breweries,
                                  beers=beers)
    return text


def main():
    args = parsed_args()
    filepath = fullpath(args.dir, 'config.toml')
    text = clean_config(filepath)
    if args.dry_run:
        print(text)
    else:
        with open(filepath, mode='w', encoding='utf8') as file_:
            file_.write(text)


if __name__ == '__main__':
    main()
