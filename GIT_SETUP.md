# Git Setup Guide

This guide will help you initialize and push this repository to git.

## Initial Setup

### 1. Initialize Git Repository

```bash
git init
```

### 2. Add All Files

```bash
git add .
```

### 3. Create Initial Commit

```bash
git commit -m "Initial commit: Module 1 - Kubernetes Architecture Deep Dive

- Complete Module 1 with 4 lessons and 4 labs
- Setup scripts for development environment and kind cluster
- Mermaid diagrams for visual learning
- Comprehensive hands-on exercises
- Testing documentation"
```

### 4. Add Remote Repository

```bash
# Replace with your actual repository URL
git remote add origin <your-repo-url>
```

### 5. Push to Remote

```bash
# Push to main branch
git branch -M main
git push -u origin main
```

## File Structure

The repository includes:

```
.
├── .gitignore          # Git ignore rules
├── .gitattributes      # Git attributes for line endings
├── LICENSE             # MIT License
├── README.md           # Main course README
├── COURSE_BUILD_PLAN.md # Course build plan
├── k8s-operators-course-syllabus.md # Course syllabus
├── GIT_SETUP.md        # This file
├── scripts/            # Setup scripts
│   ├── setup-dev-environment.sh
│   └── setup-kind-cluster.sh
└── module-01/          # Module 1 content
    ├── README.md
    ├── SUMMARY.md
    ├── TESTING.md
    ├── test-crd.sh
    ├── diagrams/
    ├── labs/
    └── lessons/
```

## What's Included

- ✅ All course content (Module 1 complete)
- ✅ Setup scripts (executable)
- ✅ Documentation
- ✅ Mermaid diagrams
- ✅ Test scripts
- ✅ .gitignore (excludes temp files, IDE files, etc.)
- ✅ .gitattributes (ensures proper line endings)

## What's Excluded (.gitignore)

- OS files (.DS_Store, Thumbs.db, etc.)
- IDE files (.vscode/, .idea/, etc.)
- Temporary files (*.tmp, *.log, etc.)
- Local test files
- Build artifacts

## Future Commits

When adding new modules, use descriptive commit messages:

```bash
git add module-02/
git commit -m "Add Module 2: Introduction to Operators

- Lesson content with Kubebuilder examples
- Hands-on labs
- Setup instructions"
```

## Branch Strategy (Optional)

Consider using branches for development:

```bash
# Create feature branch
git checkout -b module-02

# Work on module
# ... make changes ...

# Commit and push
git add .
git commit -m "Add Module 2 content"
git push -u origin module-02

# Merge to main when ready
git checkout main
git merge module-02
git push
```

## Verification

Before pushing, verify:

1. All scripts are executable: `find scripts/ -name "*.sh" -exec test -x {} \;`
2. No sensitive data in files
3. .gitignore is working: `git status` should not show temp files
4. All markdown files are properly formatted

## Notes

- The repository is ready to be pushed as-is
- All content is educational and safe to share
- No secrets or sensitive information included
- Scripts are tested and functional

