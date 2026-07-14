package superdb

func (n Node) HasSameParent(target Node) bool {
	return n.ParentID.Valid == target.ParentID.Valid &&
		n.ParentID.Int64 == target.ParentID.Int64
}
