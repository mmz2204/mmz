cd gtable
go test -tags "tablegen" -run ^TestStep1$ table/gtable
go test -tags "tablegen" -run ^TestStep2$ table/gtable
go test -tags "tablegen" -run ^TestStep3$ table/gtable -gcflags all="-l"
echo ok
pause