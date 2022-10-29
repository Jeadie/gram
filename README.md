# gram
A featured, minimal & one-line installable text editor.

Based on `antirez/kilo`

## Usage
One line install:
```bash
wget -qO - https://raw.githubusercontent.com/Jeadie/gram/main/get-gram.sh | bash
```

## Roadmap
 - Undo
 - Usage highlighting
 - Auto-save / checkpointing
 - Handle TAB more graciously
 - Handle Unicode and discrepancies between src and rende
 - Replace in file
 - On line delete, copy contents to clipboard

## Bugs
  - `jq` is not default installed on all boxes   
  - Handle string rendering with \" being ignored
  - numbers are highlighted within comments (maybe a feature?)
  - Sometimes searching spams `SEARCH: ` repeatedly. 