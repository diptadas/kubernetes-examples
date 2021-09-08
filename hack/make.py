#!/usr/bin/python
# -*- coding: utf-8 -*-

import fnmatch
import glob
import os
import os.path
import re
import subprocess


def ungroup_go_imports(*paths):
    for p in paths:
        if os.path.isfile(p):
            _ungroup_go_imports(p)
        elif os.path.isdir(p):
            for (dir, _, files) in os.walk(p):
                for f in fnmatch.filter(files, '*.go'):
                    _ungroup_go_imports(dir + '/' + f)
        else:
            for f in glob.glob(p):
                _ungroup_go_imports(f)


BEGIN_IMPORT_REGEX = ur'import \(\s*'
END_IMPORT_REGEX = ur'\)\s*'


def _ungroup_go_imports(file_name):
    print 'Ungrouping imports of file: ' + file_name
    with open(file_name, 'r+') as f:
        content = f.readlines()
        out = []
        import_block = False
        for line in content:
            c = line.strip()
            if import_block:
                if c == '':
                    continue
                elif re.match(END_IMPORT_REGEX, c) is not None:
                    import_block = False
            elif re.match(BEGIN_IMPORT_REGEX, c) is not None:
                import_block = True
            out.append(line)
        f.seek(0)
        f.writelines(out)
        f.truncate()


ungroup_go_imports('.')
subprocess.call("goimports -l -w $(find . -type f -name '*.go' -not -path './vendor/*')", shell=True)
subprocess.call('go install ./...', shell=True)
