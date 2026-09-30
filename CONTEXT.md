# TuiSnip

A terminal application for saving, organizing, and retrieving code snippets. It exists to give a SnippetsLab-like experience on Linux and macOS.

## Language

### Content

**Snippet**:
A saved, titled unit of reusable code or text, filed in at most one Folder and carrying any number of Tags. It holds one or more Fragments and an optional Description. Titles need not be unique.
_Avoid_: Entry, note, item, clip

**Fragment**:
One body of code within a Snippet, with its own Language. Every Snippet has at least one; the first version of the product allows exactly one.
_Avoid_: Body, file, part, section

**Description**:
Optional free text on a Snippet explaining what it is for. It is searched alongside the title.
_Avoid_: Notes, comment, summary

**Language**:
The programming or markup language of a Fragment. It drives syntax highlighting.
_Avoid_: Syntax, type, mode, filetype

### Organization

**Folder**:
A named node in a strict tree that files Snippets. A Snippet is in at most one Folder. Folder names need not be unique, even among siblings.
_Avoid_: Group, category, directory, collection

**Trash**:
Where deleted Snippets and Folders wait, restorable, until permanently removed.
_Avoid_: Bin, recycle bin, archive, deleted items

**Default Language**:
The Language a Folder gives each Fragment created directly within it. Every Folder has one; at the Root it is always plain text. It sets a Fragment's Language only at creation; afterwards only the user changes it.
_Avoid_: Folder language, preferred language

**Root**:
The top of the Folder tree. It is not itself a Folder; a Snippet with no Folder, and a Folder with no parent, sit at the Root.
_Avoid_: Unfiled, inbox, default folder, uncategorized

**Tag**:
A free-form label attached to any number of Snippets. A Snippet may carry any number of Tags. Tag names are case-insensitive, so `Go` and `go` are one Tag; a Tag no Snippet carries still exists until deleted.
_Avoid_: Label, keyword, category

### Retrieval

**Search**:
Finding Snippets by matching a query against their title, Description, Tags, and Fragment content: fuzzily for the first three, as a literal substring for content.
_Avoid_: Filter, find, query

**Browse**:
Navigating the Folder tree or the Tag list to reach a Snippet without typing a query.
_Avoid_: Explore, navigate

**Capture**:
Creating a new Snippet from the clipboard's current content.
_Avoid_: Import, paste, grab

**Copy**:
Placing a Fragment's content on the system clipboard.
_Avoid_: Yank, export
