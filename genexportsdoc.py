#!/usr/bin/env python3

import subprocess

cmd = ["go", "doc", "--all", "exports"]

output = subprocess.check_output(cmd).decode("utf-8")

section = [""]
package_doc = ""

in_section = True

num = 0
for out in output.splitlines():
    line = out + "\n"
    if line.startswith(" "):
        line = line[4:]
    if out.startswith("FUNCTIONS") or out.startswith("VARIABLES"):
        in_section = False
    if out.startswith("func"):
        in_section = True
        section.append(line)
        num += 1
        continue
    if in_section:
        section[num] += line


def func_name(signature: str) -> str:
    idx = signature.index("(")
    return signature[len("func ") : idx]


toc = ""
first = True

gen_sections = []
for sec in section:
    if first:
        gen_sections.append(("About the API", sec))
        first = False
        continue
    lines = sec.splitlines()

    # parse multi-line function names
    # detect end of line if we get a )
    # hacky but works for our use case
    signature_len = 1
    for line in lines:
        if not ")" in line:
            signature_len += 1
        else:
            break
    signature, doc = "\n".join(lines[:signature_len]), "\n".join(lines[signature_len:])
    body = f"Signature:\n\n```go\n{signature}\n```\n\n{doc}"
    gen_sections.append((func_name(signature), body))

data = "This document was automatically generated from the exports/exports.go file\n\n"
first = True
for title, body in gen_sections:
    if first:
        data += f"# {title}\n"
        data += f"{body}\n"
        data += "# Functions\n"
        first = False
        continue
    data += f"## {title}\n"
    data += f"{body}\n"

with open("docs/md/apidocs.md", "w+") as f:
    f.write(data)
