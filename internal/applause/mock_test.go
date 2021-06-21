package applause_test

type applauserStrub struct {
	expectedErr     error
	called          bool
	calledUserID    int
	calledMeetingID int
}

func (s *applauserStrub) Send(meetingID, uid int) error {
	s.called = true
	s.calledUserID = uid
	s.calledMeetingID = meetingID
	return s.expectedErr
}
