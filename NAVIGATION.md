# Navigation Guide

This document describes the navigation structure of the course.

## Navigation Structure

### Main Course Level
- **README.md** - Main course overview with links to all modules
- **Course Structure** - Links to each module
- **Quick Navigation** - Direct links to Module 1 lessons and labs

### Module Level
- **module-01/README.md** - Module overview with:
  - Links to all lessons
  - Links to all labs
  - Links to additional resources
  - Navigation back to course overview

### Lesson Level
Each lesson (module-01/lessons/*.md) has:
- **Header Navigation**: Links to previous/next lessons and module overview
- **Related Lab Link**: Direct link to corresponding lab
- **Footer Navigation**: Links to previous/next lessons and module overview

### Lab Level
Each lab (module-01/labs/*.md) has:
- **Header Navigation**: Links to related lesson, previous/next labs, and module overview
- **Footer Navigation**: Links to previous/next labs, related lesson, and module overview

### Reference Documents
- **SUMMARY.md** - Navigation to module overview and course overview
- **TESTING.md** - Navigation to module overview and course overview

## Navigation Patterns

### Lesson Navigation Pattern
```
[← Previous Lesson] | [Module Overview] | [Next Lesson →]
```

### Lab Navigation Pattern
```
[← Previous Lab] | [Related Lesson] | [Next Lab →]
```

### Module Navigation Pattern
```
[← Back to Course Overview] | [Next Module →]
```

## Link Verification

All navigation links have been verified:
- ✅ All lessons have navigation headers and footers
- ✅ All labs have navigation headers and footers
- ✅ Module README has links to all lessons and labs
- ✅ Main README has links to Module 1
- ✅ Reference documents have navigation
- ✅ All relative paths are correct

## Usage

Users can navigate through the course by:
1. Starting at the main README.md
2. Clicking on Module 1 to see the module overview
3. Following lesson links to read lessons
4. Following lab links to complete exercises
5. Using navigation headers/footers to move between lessons
6. Using "Related Lab" links to jump to hands-on exercises

All navigation is bidirectional - you can go forward or backward through the content.
