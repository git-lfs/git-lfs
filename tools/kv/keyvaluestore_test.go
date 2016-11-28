package kv

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreSimple(t *testing.T) {
	tmpf, err := ioutil.TempFile("", "lfstest1")
	assert.Nil(t, err)
	filename := tmpf.Name()
	defer os.Remove(filename)
	tmpf.Close()

	kvs, err := NewStore(filename)
	assert.Nil(t, err)

	// We'll include storing custom structs
	type customData struct {
		Val1 string
		Val2 int
	}
	// Needed to store custom struct
	RegisterTypeForStorage(&customData{})

	kvs.Set("stringVal", "This is a string value")
	kvs.Set("intVal", 3)
	kvs.Set("floatVal", 3.142)
	kvs.Set("structVal", &customData{"structTest", 20})

	s := kvs.Get("stringVal")
	assert.Equal(t, "This is a string value", s)
	i := kvs.Get("intVal")
	assert.Equal(t, 3, i)
	f := kvs.Get("floatVal")
	assert.Equal(t, 3.142, f)
	c := kvs.Get("structVal")
	assert.Equal(t, c, &customData{"structTest", 20})
	n := kvs.Get("noValue")
	assert.Nil(t, n)

	kvs.Remove("stringVal")
	s = kvs.Get("stringVal")
	assert.Nil(t, s)
	// Set the string value again before saving
	kvs.Set("stringVal", "This is a string value")

	err = kvs.Save()
	assert.Nil(t, err)
	kvs = nil

	// Now confirm that we can read it all back
	kvs2, err := NewStore(filename)
	assert.Nil(t, err)
	s = kvs2.Get("stringVal")
	assert.Equal(t, "This is a string value", s)
	i = kvs2.Get("intVal")
	assert.Equal(t, 3, i)
	f = kvs2.Get("floatVal")
	assert.Equal(t, 3.142, f)
	c = kvs2.Get("structVal")
	assert.Equal(t, c, &customData{"structTest", 20})
	n = kvs2.Get("noValue")
	assert.Nil(t, n)

}

func TestStoreOptimisticConflict(t *testing.T) {
	tmpf, err := ioutil.TempFile("", "lfstest2")
	assert.Nil(t, err)
	filename := tmpf.Name()
	defer os.Remove(filename)
	tmpf.Close()

	kvs1, err := NewStore(filename)
	assert.Nil(t, err)

	kvs1.Set("key1", "value1")
	kvs1.Set("key2", "value2")
	kvs1.Set("key3", "value3")
	err = kvs1.Save()
	assert.Nil(t, err)

	// Load second copy & modify
	kvs2, err := NewStore(filename)
	assert.Nil(t, err)
	// New keys
	kvs2.Set("key4", "value4_fromkvs2")
	kvs2.Set("key5", "value5_fromkvs2")
	// Modify a key too
	kvs2.Set("key1", "value1_fromkvs2")
	err = kvs2.Save()
	assert.Nil(t, err)

	// Now modify first copy & save; it should detect optimistic lock issue
	// New item
	kvs1.Set("key10", "value10")
	// Overlapping item; since we save second this will overwrite one from kvs2
	kvs1.Set("key4", "value4")
	err = kvs1.Save()
	assert.Nil(t, err)

	// This should have merged changes from kvs2 in the process
	v := kvs1.Get("key1")
	assert.Equal(t, "value1_fromkvs2", v) // this one was modified by kvs2
	v = kvs1.Get("key2")
	assert.Equal(t, "value2", v)
	v = kvs1.Get("key3")
	assert.Equal(t, "value3", v)
	v = kvs1.Get("key4")
	assert.Equal(t, "value4", v) // we overwrote this so would not be merged
	v = kvs1.Get("key5")
	assert.Equal(t, "value5_fromkvs2", v)

}

func TestStoreReduceSize(t *testing.T) {
	tmpf, err := ioutil.TempFile("", "lfstest3")
	assert.Nil(t, err)
	filename := tmpf.Name()
	defer os.Remove(filename)
	tmpf.Close()

	kvs, err := NewStore(filename)
	assert.Nil(t, err)

	kvs.Set("key1", "I woke up in a Soho doorway")
	kvs.Set("key2", "A policeman knew my name")
	kvs.Set("key3", "He said 'You can go sleep at home tonight")
	kvs.Set("key4", "If you can get up and walk away'")

	assert.NotNil(t, kvs.Get("key1"))
	assert.NotNil(t, kvs.Get("key2"))
	assert.NotNil(t, kvs.Get("key3"))
	assert.NotNil(t, kvs.Get("key4"))

	assert.Nil(t, kvs.Save())

	stat1, _ := os.Stat(filename)

	// Remove all but 1 key & save smaller version
	kvs.Remove("key2")
	kvs.Remove("key3")
	kvs.Remove("key4")
	assert.Nil(t, kvs.Save())

	// Now reload fresh & prove works
	kvs = nil

	kvs, err = NewStore(filename)
	assert.Nil(t, err)
	assert.NotNil(t, kvs.Get("key1"))
	assert.Nil(t, kvs.Get("key2"))
	assert.Nil(t, kvs.Get("key3"))
	assert.Nil(t, kvs.Get("key4"))

	stat2, _ := os.Stat(filename)

	assert.True(t, stat2.Size() < stat1.Size(), "Size should have reduced, was %d now %d", stat1.Size(), stat2.Size())

}
