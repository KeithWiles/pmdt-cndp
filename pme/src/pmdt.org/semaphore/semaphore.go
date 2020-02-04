// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2019-2020 Intel Corporation

package semaphore

/*
#include <fcntl.h>
#include <sys/stat.h>
#include <semaphore.h>
#include <stdlib.h>
#include <errno.h>

typedef struct {
	sem_t *sem;
	int err;
} sema_t;

int
_sem_init(sem_t *sem, int shared, int initValue)
{
	int r;

	r = sem_init(sem, shared, initValue);

	return (r == 0)? 0 : errno;
}

int
_sem_destroy(sem_t *s)
{
	int r;

	r = sem_destroy(s);

	return (r == 0)? 0 : errno;
}

sem_t *
_sem_open(char *name, int flags, mode_t mode, unsigned int val)
{
	sem_t *s;

	s = sem_open(name, flags, mode, val);

	return (s == SEM_FAILED)? NULL : s;
}

int
_sem_close(sem_t *s)
{
	int r;

	r = sem_close(s);

	return (r == 0)? 0 : errno;
}

int
_sem_post(sem_t *s)
{
	int r;

	r = sem_post(s);

	return (r == 0)? 0 : errno;
}

int _sem_wait(sem_t *s) {
	int r;

	r = sem_wait(s);

	return (r == 0) ? 0 : errno;
}

int _sem_trywait(sem_t *s) {
	int r;

	r = sem_trywait(s);

    return (r == 0) ? 0 : errno;
}

int
_sem_unlink(char* name) {
	int r;

	r = sem_unlink((const char*) name);

	return (r == 0) ? 0 : errno;
}

const struct timespec *
new_timespec(unsigned int sec, unsigned int nsec) {
	struct timespec* val = (struct timespec*)malloc(sizeof(struct timespec));

    val->tv_sec = sec;
	val->tv_nsec = nsec;

    return (const struct timespec*)val;
}

void
free_timespec(struct timespec *ts)
{
	free(ts);
}

int
_sem_timedwait(sem_t *s, const struct timespec *ts)
{
	return sem_timedwait(s, ts);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Semaphore the Go version of the semaphore structure
type Semaphore struct {
	name string
	sema *C.sem_t
}

func (s *Semaphore) isValid() bool {
	if s == nil || s.sema == nil {
		return false
	}
	return true
}

// Init to create a semaphore from a memory location
func Init(addr []byte, shared, val int) (*Semaphore, error) {

	s := (*C.sem_t)(unsafe.Pointer(&addr[0]))

	sema := &Semaphore{name: "init_sem"}

	r := C._sem_init(s, C.int(shared), C.int(val))
	if r != 0 {
		return nil, fmt.Errorf("failed to initialize semaphore")
	}

	sema.sema = s

	return sema, nil
}

// Open a semaphore using the given name
func Open(name string, lock bool, mode int32, initVal uint32) (*Semaphore, error) {
	name = fmt.Sprintf("/%s", name)
	cName := C.CString(name)

	flags := C.O_CREAT
	if lock {
		flags = flags | C.O_EXCL
	}

	s, err := C._sem_open(cName, C.int(flags), C.mode_t(mode), C.uint(initVal))
	C.free(unsafe.Pointer(cName))
	if err != nil {
		return nil, fmt.Errorf("error opening semaphore: %v", err)
	}

	sem := &Semaphore{
		name: name,
		sema: s,
	}

	return sem, nil
}

// Close the semaphore
func (s *Semaphore) Close() error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	err := C._sem_close(s.sema)
	if err != 0 {
		return fmt.Errorf("error closing semaphore")
	}
	return nil
}

// Post to a semaphore
func (s *Semaphore) Post() error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	err := C._sem_post(s.sema)
	if err == 0 {
		return nil
	}
	return fmt.Errorf("post to semaphore failed")
}

// Value of the semaphore
func (s *Semaphore) Value() (int, error) {
	var val C.int

	if !s.isValid() {
		return 0, fmt.Errorf("semaphore is not valid")
	}

	ret, err := C.sem_getvalue(s.sema, &val)
	if ret != 0 {
		return int(ret), err
	}

	return int(val), nil
}

// Wait on a semaphore
func (s *Semaphore) Wait() error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	err := C._sem_wait(s.sema)
	if err == 0 {
		return nil
	}
	return fmt.Errorf("error waiting on semaphore")
}

// TryWait on a semaphore
func (s *Semaphore) TryWait() error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	r := C._sem_trywait(s.sema)
	if r == 0 || r == C.EAGAIN {
		return nil
	}
	return fmt.Errorf("error trying wait on semaphore")
}

// TimedWait on a semaphore
func (s *Semaphore) TimedWait(sec, nano uint64) error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	ts := C.new_timespec(C.uint(sec), C.uint(nano))

	r := C._sem_timedwait(s.sema, ts)
	if r != 0 {
		return fmt.Errorf("sem_timedwait() failed")
	}
	return nil
}

// Unlink a semaphore file
func (s *Semaphore) Unlink() error {

	if !s.isValid() {
		return fmt.Errorf("semaphore is not valid")
	}

	cName := C.CString(fmt.Sprintf("/%s", s.name))

	err := C._sem_unlink(cName)
	C.free(unsafe.Pointer(cName))
	if err == 0 {
		return nil
	}

	return fmt.Errorf("error unlinking semaphore")
}
