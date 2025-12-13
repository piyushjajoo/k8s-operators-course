#!/bin/bash
# Script to add Jekyll front matter to all lesson and lab files
# Run from the root of the k8s-operators-course directory

set -e

get_module_title() {
    case "$1" in
        "module-01") echo "Module 1: Kubernetes Architecture" ;;
        "module-02") echo "Module 2: Introduction to Operators" ;;
        "module-03") echo "Module 3: Building Custom Controllers" ;;
        "module-04") echo "Module 4: Advanced Reconciliation" ;;
        "module-05") echo "Module 5: Webhooks & Admission Control" ;;
        "module-06") echo "Module 6: Testing & Debugging" ;;
        "module-07") echo "Module 7: Production Considerations" ;;
        "module-08") echo "Module 8: Advanced Topics" ;;
        *) echo "Unknown Module" ;;
    esac
}

add_frontmatter() {
    local file="$1"
    local title="$2"
    local nav_order="$3"
    local parent="$4"
    local permalink="$5"

    # Check if file already has front matter
    if head -1 "$file" | grep -q "^---"; then
        echo "Skipping $file (already has front matter)"
        return
    fi

    local content=$(cat "$file")
    
    cat > "$file" << EOF
---
layout: default
title: "$title"
nav_order: $nav_order
parent: "$parent"
grand_parent: Modules
permalink: $permalink
mermaid: true
---

$content
EOF
    echo "Added front matter to $file"
}

capitalize_words() {
    echo "$1" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1'
}

# Process each module
for module in module-01 module-02 module-03 module-04 module-05 module-06 module-07 module-08; do
    if [ ! -d "$module" ]; then
        continue
    fi
    
    parent_title=$(get_module_title "$module")
    module_num="${module#module-}"
    
    # Process lessons
    if [ -d "$module/lessons" ]; then
        for lesson in "$module/lessons"/*.md; do
            if [ -f "$lesson" ]; then
                filename=$(basename "$lesson" .md)
                # Extract lesson number from filename (e.g., 01-control-plane -> 1)
                lesson_order=$(echo "$filename" | cut -d'-' -f1 | sed 's/^0*//')
                if [ -z "$lesson_order" ]; then
                    lesson_order=1
                fi
                
                # Create title from filename
                lesson_name=$(echo "$filename" | sed 's/^[0-9]*-//')
                lesson_name_formatted=$(capitalize_words "$lesson_name")
                title="${module_num}.${lesson_order} ${lesson_name_formatted}"
                
                permalink="/${module}/lessons/${lesson_name}/"
                
                add_frontmatter "$lesson" "$title" "$lesson_order" "$parent_title" "$permalink"
            fi
        done
    fi
    
    # Process labs  
    if [ -d "$module/labs" ]; then
        for lab in "$module/labs"/*.md; do
            if [ -f "$lab" ]; then
                filename=$(basename "$lab" .md)
                # Extract lab number
                lab_num=$(echo "$filename" | sed 's/lab-0*//' | cut -d'-' -f1)
                if [ -z "$lab_num" ]; then
                    lab_num=1
                fi
                nav_order=$((10 + lab_num))
                
                # Create title from filename
                lab_name=$(echo "$filename" | sed 's/^lab-[0-9]*-//')
                lab_name_formatted=$(capitalize_words "$lab_name")
                title="Lab ${module_num}.${lab_num}: ${lab_name_formatted}"
                
                permalink="/${module}/labs/${lab_name}/"
                
                add_frontmatter "$lab" "$title" "$nav_order" "$parent_title" "$permalink"
            fi
        done
    fi
done

echo ""
echo "✅ Front matter added to all lesson and lab files!"
echo ""
echo "To preview your site locally:"
echo "  bundle install"
echo "  bundle exec jekyll serve"
echo ""
echo "Then open http://localhost:4000/k8s-operators-course/"
