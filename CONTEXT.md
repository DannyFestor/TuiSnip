# TuiSnip

A terminal application for saving, organizing, and retrieving code snippets. It exists to give a SnippetsLab-like experience on Linux and macOS.

## Language

### Content

**Snippet**:
A saved, titled unit of reusable code or text, filed in at most one Folder and carrying any number of Tags. It holds one or more Fragments. Titles need not be unique.
_Avoid_: Entry, note, item, clip

**Fragment**:
One body of code within a Snippet, with its own Language. Every Snippet has at least one; the first version of the product allows exactly one.
_Avoid_: Body, file, part, section

**Language**:
The programming or markup language of a Fragment. It drives syntax highlighting.
_Avoid_: Syntax, type, mode, filetype

### Organization

**Folder**:
A named node in a strict tree that files Snippets. A Snippet is in at most one Folder.
_Avoid_: Group, category, directory, collection

**Root**:
The top of the Folder tree. It is not itself a Folder; a Snippet with no Folder, and a Folder with no parent, sit at the Root.
_Avoid_: Unfiled, inbox, default folder, uncategorized

**Tag**:
A free-form label attached to any number of Snippets. A Snippet may carry any number of Tags.
_Avoid_: Label, keyword, category

### Retrieval

**Search**:
Finding Snippets by fuzzy match on their title, Fragment content, and Tags.
_Avoid_: Filter, find, query

**Browse**:
Navigating the Folder tree or the Tag list to reach a Snippet without typing a query.
_Avoid_: Explore, navigate

**Copy**:
Placing a Fragment's content on the system clipboard.
_Avoid_: Yank, export
