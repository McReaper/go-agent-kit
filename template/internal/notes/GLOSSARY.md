# The notes module's words

Each type of the domain, what it means here, and what makes one.

- **Note**: a short text a person or an agent keeps: its title, its body, when it was made, and its id. Made by `New`.
- **Title**: what a note is called: trimmed, 1 to 80 characters, with at least one ASCII letter or digit once accents are dropped, to name the note by. Made by `ParseTitle`.
- **NoteID**: the name a note is stored and found under: its title's slug and the UTC second it was made (`buy-milk-20261030t220509`). A second note with the same title in the same second gets the same id, and adding it is refused. Made by `New`; `ParseNoteID` checks one read back: lowercase letters and digits joined by single hyphens.
