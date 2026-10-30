package notes

type NotesModule struct{} // want `M2: notes.NotesModule repeats the package name: call it notes.Module`

type Notes struct{}

type Notebook struct{}

func NotesFrom(s string) Notes { return Notes{} } // want `M2: notes.NotesFrom repeats the package name`

var notesCount int
