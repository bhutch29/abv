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
import os
import sys
import unicodedata

from argparse import ArgumentParser, RawTextHelpFormatter
from os.path import expanduser, realpath

import toml


def fullpath(*args):
    return realpath(os.path.join(*args))


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


def unpack_table(lines):
    table_toml = toml.loads(''.join(lines))
    name = list(table_toml)[0]
    # Format table title
    title = '[{name}]\n'.format(name=name)
    # Define a whitespace insertion method based on the longest key
    longest = max([len(key) for key in table_toml[name]])
    def align(str_):
        return ' '*(longest-len(str_)+1)
    # Lexicographically sort and whitespace-align table items
    table = sorted(table_toml[name].items(), key=lambda x: x[0])
    line = '"{key}"{ws}= "{val}"'
    items = '\n'.join([line.format(key=key, ws=align(key), val=val) for key, val in table])
    return title + items + '\n\n'


def parse_toml(lines):
    final = []
    inside_table = False
    for line in lines:
        if line.strip().startswith('#'):
            if not inside_table:
                final.append(line)
            continue
        elif line.strip().startswith('['):
            if inside_table:
                final.append(unpack_table(table_lines))
            inside_table = True
            table_lines = [line]
            continue
        elif line.strip().startswith('"') and inside_table:
            table_lines.append(line)
            continue
        elif line.strip() == '':
            if not inside_table:
                final.append(line)
            continue
        else:
            inside_table = False
            final.append(line)
    if inside_table:
        final.append(unpack_table(table_lines))
    text = ''.join(final)
    return text.strip() + '\n'


def clean_config(filepath):
    with open(filepath, mode='r', encoding='utf8') as file_:
        lines = file_.readlines()
    text = parse_toml(lines)
    text = unicodedata.normalize('NFC', text)
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
