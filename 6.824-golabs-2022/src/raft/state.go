package raft

import "time"

type State int

const (
	Follower  State = 0
	Candidate State = 1
	Leader    State = 2
)

func (rf *Raft) toFollower() {
	rf.state = Follower
}

func (rf *Raft) toCandidate() {
	rf.state = Candidate
	rf.votedFor = rf.me
}

func (rf *Raft) toLeader() {
	rf.state = Leader
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := range rf.peers {
		rf.nextIndex[i] = rf.commitIndex + 1
	}

	go rf.updateCommitIndex()
	// go rf.Start(nil)
	// 立即开始心跳
	rf.HeartBeatTimer.Reset(time.Duration(0))
}

// updateCommitIndex 5.3 启动协程，每次日志复制时都更新Index，等大多数日志的matchIndex[i] >= N 时，CommitIndex + 1
func (rf *Raft) updateCommitIndex() {
	for !rf.killed() {
		rf.mu.Lock()
		time.Sleep(time.Duration(10) * time.Millisecond)
		for i := rf.LastLog().Index; i > rf.commitIndex; i-- {
			num := 0
			for j := range rf.peers {
				if j == rf.me {
					continue
				}
				if rf.matchIndex[j] >= i && rf.currentTerm == rf.log[i].Term {
					num++
				}
			}

			if num > len(rf.peers)/2 {
				rf.commitIndex = i
				rf.applyCond.Broadcast()
				break
			}
		}

		rf.mu.Unlock()
	}
}