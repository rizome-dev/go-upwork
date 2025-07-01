#!/usr/bin/env python3
import os
import re
from pathlib import Path

def split_docs(input_file="DOCS.md", output_dir="docs_split"):
    """Split the large DOCS.md file into smaller section-based files."""
    
    # Create output directory
    Path(output_dir).mkdir(exist_ok=True)
    
    with open(input_file, 'r', encoding='utf-8') as f:
        content = f.read()
    
    lines = content.split('\n')
    
    # Track sections and their content
    sections = []
    current_section = {'title': 'Introduction', 'content': [], 'level': 1}
    
    # Patterns to identify sections
    h1_pattern = re.compile(r'^# (.+)$')
    h2_pattern = re.compile(r'^## (.+)$')
    h3_pattern = re.compile(r'^### (.+)$')
    h4_pattern = re.compile(r'^#### (.+)$')
    h5_pattern = re.compile(r'^##### (.+)$')
    
    for line in lines:
        # Check for headers
        h1_match = h1_pattern.match(line)
        h2_match = h2_pattern.match(line)
        h3_match = h3_pattern.match(line)
        h4_match = h4_pattern.match(line)
        h5_match = h5_pattern.match(line)
        
        # If we find a major section (H1 or H2), save the current section and start a new one
        if h1_match or h2_match:
            if current_section['content']:
                sections.append(current_section)
            
            if h1_match:
                title = h1_match.group(1)
                level = 1
            else:
                title = h2_match.group(1)
                level = 2
                
            current_section = {
                'title': title,
                'content': [line],
                'level': level
            }
        else:
            current_section['content'].append(line)
    
    # Don't forget the last section
    if current_section['content']:
        sections.append(current_section)
    
    # Write sections to files
    section_index = {}
    for i, section in enumerate(sections):
        # Clean filename
        filename = re.sub(r'[^\w\s-]', '', section['title'])
        filename = re.sub(r'[-\s]+', '-', filename)
        filename = filename.strip('-').lower()
        
        # Handle empty or very short filenames
        if not filename or len(filename) < 3:
            filename = f"section_{i:03d}"
        
        filepath = os.path.join(output_dir, f"{i:03d}_{filename}.md")
        
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write('\n'.join(section['content']))
        
        section_index[section['title']] = {
            'file': filepath,
            'lines': len(section['content']),
            'level': section['level']
        }
        
        print(f"Created: {filepath} ({len(section['content'])} lines)")
    
    # Create an index file
    index_path = os.path.join(output_dir, "00_index.md")
    with open(index_path, 'w', encoding='utf-8') as f:
        f.write("# Documentation Index\n\n")
        f.write(f"Total sections: {len(sections)}\n\n")
        
        for title, info in section_index.items():
            indent = "  " * (info['level'] - 1)
            f.write(f"{indent}- [{title}](./{os.path.basename(info['file'])}) ({info['lines']} lines)\n")
    
    print(f"\nCreated index at: {index_path}")
    print(f"Total sections: {len(sections)}")
    
    # Also create a special file for large code blocks or schemas if they exist
    extract_schemas_and_types(content, output_dir)

def extract_schemas_and_types(content, output_dir):
    """Extract GraphQL schemas, types, and large code blocks into separate files."""
    
    # Pattern to find GraphQL type definitions and schemas
    type_pattern = re.compile(r'```(?:graphql|typescript|javascript)?\n(type\s+\w+.*?)```', re.DOTALL)
    schema_pattern = re.compile(r'```(?:graphql|json)?\n(\{[\s\S]*?\})```', re.DOTALL)
    
    types = type_pattern.findall(content)
    schemas = schema_pattern.findall(content)
    
    if types:
        types_path = os.path.join(output_dir, "graphql_types.md")
        with open(types_path, 'w', encoding='utf-8') as f:
            f.write("# GraphQL Types\n\n")
            for i, type_def in enumerate(types):
                f.write(f"## Type {i+1}\n\n```graphql\n{type_def}```\n\n")
        print(f"Extracted {len(types)} GraphQL types to: {types_path}")
    
    if schemas:
        schemas_path = os.path.join(output_dir, "schemas.md")
        with open(schemas_path, 'w', encoding='utf-8') as f:
            f.write("# Schemas\n\n")
            for i, schema in enumerate(schemas):
                if len(schema) > 100:  # Only save large schemas
                    f.write(f"## Schema {i+1}\n\n```json\n{schema}```\n\n")
        print(f"Extracted schemas to: {schemas_path}")

if __name__ == "__main__":
    split_docs()