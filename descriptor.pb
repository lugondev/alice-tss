
�

ping.protopb"
PingRequest"%
	PongReply
message (	Rmessage"

MsgRequest"$
MsgReply
message (	RmessageBZpb/bproto3
�
service.protopb
ping.proto2W
Ping(
Ping.pb.PingRequest.pb.PongReply" %
Msg.pb.MsgRequest.pb.MsgReply" BZpb/bproto3
�
	tss.protopb"

DKGRequest"S
SignRequest
hash (	Rhash
pubkey (	Rpubkey
message (	Rmessage"<
ReshareRequest
hash (	Rhash
pubkey (	Rpubkey"B
RVSignatureReply
r (	Rr
s (	Rs
hash (	Rhash"l
DkgReply
x (	Rx
y (	Ry
pubkey (	Rpubkey
address (	Raddress
hash (	Rhash"Q
CheckSignatureByPubkeyRequest
message (	Rmessage
pubkey (	Rpubkey"
ServiceReply2�

TssService6
SignMessage.pb.SignRequest.pb.RVSignatureReply" -
RegisterDKG.pb.DKGRequest.pb.DkgReply" 1
Reshare.pb.ReshareRequest.pb.ServiceReply" BZpb/bproto3