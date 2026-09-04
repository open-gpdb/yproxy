package message

import (
	"encoding/binary"
)

// CleanupMessage requests removal of an object or garbage collection.
// Garbage collection uses soft deletion by default; CrazyDrop switches it to
// hard deletion. A single-file request is handled by the storage deleter.
type CleanupMessage struct { // Seg port
	Name      string // File path
	Port      uint64 // Port segment/instance DB
	Segnum    uint64 // Segment number
	Confirm   bool   // Execute or Dry-run
	Garbage   bool   // Run vacuum-style garbage deletion instead of deleting a single file
	CrazyDrop bool   // For garbage mode: delete immediately instead of moving to trash
}

var _ ProtoMessage = &CleanupMessage{}

func BuildCleanupMessage(name string, port uint64, seg uint64, confirm bool, garbage bool) *CleanupMessage {
	return &CleanupMessage{
		Name:    name,
		Port:    port,
		Segnum:  seg,
		Confirm: confirm,
		Garbage: garbage,
	}
}

func (c *CleanupMessage) Encode() []byte {
	bt := []byte{
		byte(MessageTypeDelete),
		0,
		0,
		0,
	}

	if c.Confirm {
		bt[1] = 1
	}
	if c.Garbage {
		bt[2] = 1
	}
	if c.CrazyDrop {
		bt[3] = 1
	}

	bt = append(bt, []byte(c.Name)...)
	bt = append(bt, 0)

	p := make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(c.Port))
	bt = append(bt, p...)

	p = make([]byte, 8)
	binary.BigEndian.PutUint64(p, uint64(c.Segnum))
	bt = append(bt, p...)

	ln := len(bt) + 8
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, uint64(ln))
	return append(bs, bt...)
}

func (c *CleanupMessage) Decode(body []byte) {
	if body[1] == 1 {
		c.Confirm = true
	}
	if body[2] == 1 {
		c.Garbage = true
	}
	if body[3] == 1 {
		c.CrazyDrop = true
	}
	c.Name, _ = GetCstring(body[4:])
	c.Port = binary.BigEndian.Uint64(body[len(body)-16 : len(body)-8])
	c.Segnum = binary.BigEndian.Uint64(body[len(body)-8:])
}
