/*
 * Implementation of the client in Path ORAM by Stefanov et. al.
 */
package oram

import (
	"crypto/rand"
	"errors"
	"math"
	// "net"
    "fmt"
    "encoding/json"
    "log"
)

/*
 * The client
 */
type Client struct {
	N          int
	L          int
	B          int
	Z          int
	stash      []Block
	pos        map[int]int
	keys       []byte
    conn       *Conn
}

/*
 * Initialize a Client
 *
 * Returns a new Client with params:
 *   N: number of blocks outsourced to server
 *   L: height of the tree
 *   B: Number of bytes in each block (fixed to 32 bytes)
 *   Z: Number of blocks in each bucket
 *   S: size of client's stash in blocks
 *   key: 16-byte key to encrypt blocks with
 */
func InitClient(N int, Z int, B int) *Client {
	c := &Client{N: N, B: B, Z: Z, pos: make(map[int]int)}

	// initialize pos map as random values
	// create cryptographically secure shuffling of leaves
	c.L = int(math.Ceil(math.Log2(float64(N))))
	random_leaves := random_perm(1 << uint(c.L))

	// assign each block with a unique random leaf
	for i := 0; i < N; i++ {
		// not all leaves will be used if N < 2^L but that's okay
		c.pos[i] = int(random_leaves[i])
	}

	// init stash map
	//c.stash = make(map[string][]Block)

	// initialize empty server map and keys map
	//c.keys = make(map[string][]byte)

	return c
}

func (c *Client) AddServer(serverAddr string, N int, Z int) error {
	// add new server
    var err error
    c.conn, err = OpenConnection(serverAddr)
    if err != nil {
        fmt.Println("Server not up")
    }
    //InitServer(N, Z,)

	// generate random key for that server
	key := make([]byte, 16)
	rand.Read(key)
	c.keys = key

	// init stash
	S := c.N * c.L
	c.stash = make([]Block, 0, S)

	// initialize serverside storage as all dummy blocks
	err = c.init_server_storage(key)

	return err
}

func (c *Client) RemoveServer() {
    CloseConnection(c.conn)
}

func (c *Client) init_server_storage(key []byte) error {
	// encrypt c.Z dummy blocks to get a bucket, and write to every node in tree
	for i := 0; i <= c.L; i++ {
		for j := 0; j < (1 << uint(i)); j++ {
			bucket := make_bucket(nil, c.Z, key, c.B)
			c.RemoteWrite(bucket, i, j)
		}
	}

	return nil
}

// data must be of size B
func (c *Client) Access(write bool, a int, data []byte) ([]byte, error) {
    ret := make([]byte, c.B)

	// get encryption key for this server
	key := c.keys

	// get position from posmap
	x, prs := c.pos[a]
	if prs == false {
		return ret, errors.New("Tried to look up invalid block number in pos!")
	}

	// map block a to new random leaf
	num_leaves := 1 << uint(c.L)
	new_leaf := gen_int(num_leaves)
	c.pos[a] = new_leaf

	// read path containing block a (i.e. the path to leaf x)
	buckets, err := c.get_path_buckets(x)
	if err != nil {
		return ret, err
	}

	// write nondummy blocks into stash, record which indexes they're at
	cur_stash := c.stash
	nondummy := find_nondummy(buckets, key)
	path_start := len(cur_stash)
	cur_stash = append(cur_stash, nondummy...)
	path_end := len(cur_stash)
	// fmt.Println("path_start:", path_start)
	// fmt.Println("path_end:", path_end)

	// fmt.Println("nondummy blocks:", nondummy)

	// find index of block we're looking for
	i := slice_find_block(cur_stash, a)

	// modify contents of block in the stash for a write operation
	if write {
		// fmt.Println("if writing, found elem at idx", i)
		new_blk := data
		// if element not found, add it as a stash block
		if i == -1 {
			cur_stash = append(cur_stash, new_blk)
		} else {
			cur_stash[i] = new_blk
		}

		ret = data
	} else {
		if i != -1 {
			ret = cur_stash[i]
		}
	}

	//fmt.Printf("Calculating writeback took %s\n", elapsed)


	// fmt.Println("current stash after writing nondummy blocks:", cur_stash)

	// find intersections between old and new path
	old_path, err := c.get_path(x)
	if err != nil {
		return ret, err
	}
	// fmt.Println("old path:", old_path)

	new_path, err := c.get_path(new_leaf)
	if err != nil {
		return ret, err
	}
	// fmt.Println("new path:", new_path)

	num_inters := 0
	for i := range old_path {
		if old_path[i] == new_path[i] {
			num_inters += 1
		}
	}

	// fill list of intersections greedily (starting from the leaf)
	inters := make([]int, num_inters)
	for i := 0; i < num_inters; i++ {
		inters[i] = old_path[num_inters-1-i]
	}
	// fmt.Println("Found intersection:", inters)

	// write back path

	// write back nondummy blocks first
	blk_to_write := make([]Block, 0, len(inters))
	num := cap(blk_to_write) - len(blk_to_write)
	if path_end - path_start < num  {
		num = path_end - path_start
	}

	// fmt.Println("len of blk to write:", len(blk_to_write))
	// fmt.Println("cap of blk to write:", cap(blk_to_write))

	blk_to_write = append(blk_to_write, cur_stash[path_start : path_start + num]...)
	// fmt.Println("Writing back nondummy blocks:", blk_to_write)

	// if there aren't enough blocks to write back, then write stash blocks
	if len(blk_to_write) < cap(blk_to_write) {
		num = cap(blk_to_write) - len(blk_to_write)
		if len(cur_stash) < num {
			num = len(cur_stash)
		}

		blk_to_write = append(blk_to_write, cur_stash[:num]...)
		cur_stash = append(cur_stash[num:])
		// fmt.Printf("Appended %d stash block...\n", num)
	}

	// if out of stash blocks, write dummy blocks
	if len(blk_to_write) < cap(blk_to_write) {
		dummy := make([]Block, len(inters)-len(blk_to_write))
		for i := range dummy {
			dummy[i] = dummy_block(c.B)
		}

		blk_to_write = append(blk_to_write, dummy...)
		// fmt.Printf("Appending %d dummy blocks...\n", len(dummy))
	}

	// write blocks
	for i := range inters {
		cur_l := len(inters) - 1 - i
		bucket := make_bucket([]Block{blk_to_write[i]}, c.Z, key, c.B)
		c.RemoteWrite(bucket, cur_l, inters[i])
	}

	// TODO: calculate which blocks to write back into the original path?

	// if can't write anymore blocks, copy to the stash
	if len(inters) < len(blk_to_write) {
		extra_blks := blk_to_write[len(inters):]
		cur_stash = append(cur_stash, extra_blks...)

		// fmt.Printf("Adding %d blocks to stash...\n", len(extra_blks))
	}

	c.stash = cur_stash

	return ret, nil
}

// returns the path to block n
func (c *Client) get_path(n int) ([]int, error) {
    if n < 0 || n >= c.N {
        // block not found
        return nil, errors.New("Block number out of range")
    }

    // for each level of the tree, get which index the bucket is
    path := make([]int, c.L+1)

    cur_n := n
    for i := c.L; i >= 0; i-- {
        path[i] = cur_n
        cur_n /= 2
    }

    return path, nil
}

// access the files stored on disk to retrieve buckets of a path
func (c *Client) get_path_buckets(n int) ([]Bucket, error) {
    path, err := c.get_path(n)
    if err != nil {
        return nil, err
    }

    bux := make([]Bucket, c.L+1)
    for i := range bux {
        bucket, err := c.RemoteRead(i, path[i])
        if err != nil {
            return nil, err
        }

        bux[i] = bucket
    }

    return bux, nil
}

func (c *Client) RemoteRead(l int, n int) (Bucket, error) {
    req := &ReadRequest{
        L: l,
        N: n,
    }
    resp := &ReadResponse{}
    var respErr error
    SendMessageWithConnection(c.conn, READ, req, resp, &respErr)
    return resp.BucketResp, nil
}

func (c *Client) RemoteWrite(b Bucket, l int, n int) {
    req := &WriteRequest{
        L: l,
        N: n,
        BucketReq: b,
    }
    SendMessageWithConnectionNoResp(c.conn, WRITE, req)
}

func (c *Client) SaveState() {
    stashCts := make([][]byte, 0)
    for i := 0; i < len(c.stash); i++ {
        stashCt := encrypt(c.stash[i], c.keys)
        stashCts = append(stashCts, stashCt)
    }
    posMapBuf, err := json.Marshal(c.pos)
    if err != nil {
        log.Fatal("Failed to marshal position map: ", err)
    }
    posMapCt := encrypt(posMapBuf, c.keys)
    req := &SaveRequest{
        StashCts: stashCts,
        PosMapCt: posMapCt,
    }
    SendMessageWithConnectionNoResp(c.conn, SAVE, req)
}

func (c *Client) LoadState() {
    req := &LoadRequest{}
    resp := &LoadResponse{}
    var respErr error
    SendMessageWithConnection(c.conn, LOAD, req, resp, &respErr)
    posMapBuf := decrypt(resp.PosMapCt, c.keys)
    err := json.Unmarshal(posMapBuf, &c.pos)
    if err != nil {
        log.Fatal("Failed to unmarshal position map: ", err)
    }
    c.stash = make([]Block, 0)
    for i := 0; i < len(resp.StashCts); i++ {
        stashBuf := decrypt(resp.StashCts[i], c.keys)
        c.stash = append(c.stash, stashBuf)
    }
}

func (c *Client) CommitToReq(cm []byte) {
    ct := encrypt(cm, c.keys)
    req := &CommitRequest{Cm: ct}
    SendMessageWithConnectionNoResp(c.conn, COMMIT, req)
}
