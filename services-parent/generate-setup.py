#!/usr/bin/env python3
import sys
import os

# Read template
template_path = sys.argv[1]
output_path = sys.argv[2]
package_name = sys.argv[3]
package_version = sys.argv[4]
artifact_id = sys.argv[5]

with open(template_path, 'r') as f:
    content = f.read()

# Replace placeholders
content = content.replace('${python.package.name}', package_name)

# Normalize version for Python (convert SNAPSHOT to dev0)
normalized_version = package_version.replace('-SNAPSHOT', '.dev0')
content = content.replace('${python.package.version}', normalized_version)

content = content.replace('${project.artifactId}', artifact_id)

# Write output
with open(output_path, 'w') as f:
    f.write(content)

print(f"Generated setup.py at {output_path}")
