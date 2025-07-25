package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/echotools/nevr-common/v3/gameapi"
	"github.com/echotools/nevr-common/v3/rtapi"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	protojsonMarshaler = &protojson.MarshalOptions{
		UseProtoNames:   false,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}
	protojsonUnmarshaler = &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
)

func TestGameAPISessionCompression(t *testing.T) {
	// Open the echoreplay file as a zip
	fn := "rec_2025-05-29_18-14-10.echoreplay"

	file, err := os.Open(fn)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", fn, err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", fn, err)
	}
	zipReader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		t.Fatalf("failed to create zip reader for file %s: %v", fn, err)
	}
	// Find the GameAPISession file in the zip
	var gameAPISessionFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == fn {
			gameAPISessionFile = f
			break
		}
	}
	if gameAPISessionFile == nil {
		t.Fatalf("not found in zip file %s", fn)
	}
	// Open the GameAPISession file
	fileReader, err := gameAPISessionFile.Open()
	if err != nil {
		t.Fatalf("failed to open GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}
	// Read the file content
	data, err := io.ReadAll(fileReader)
	if err != nil {
		t.Fatalf("failed to read GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}
	// Close the file reader
	if err := fileReader.Close(); err != nil {
		t.Fatalf("failed to close GameAPISession file %s: %v", gameAPISessionFile.Name, err)
	}

	frames := &[]*rtapi.SessionUpdateMessage{}
	// Unmarshal the line into a GameAPISession
	if err := UnmarshalReplay(data, frames); err != nil {
		t.Fatalf("failed to unmarshal GameAPISession line: %v", err)
	}

	outputFile, err := os.Create("GameAPISession_protobuf.bin")
	if err != nil {
		t.Fatalf("failed to create output file GameAPISession_protobuf.bin: %v", err)
	}
	defer outputFile.Close()
	// Create a protobuf encoder
	t.Logf("Encoding %d frames to protobuf", len(*frames))
	for _, frame := range *frames {
		// Encode the GameAPISession to protobuf
		protbufEncodedData, err := proto.Marshal(frame)
		if err != nil {
			t.Fatalf("failed to marshal GameAPISession to protobuf: %v", err)
		}
		t.Logf("Encoded GameAPISession to protobuf, size: %d bytes", len(protbufEncodedData))
		// Write the protobuf encoded data to the output file
		if _, err := outputFile.Write(protbufEncodedData); err != nil {
			t.Fatalf("failed to write GameAPISession_protobuf.bin: %v", err)
		}
	}
	// Close the output file
	if err := outputFile.Close(); err != nil {
		t.Fatalf("failed to close output file GameAPISession_protobuf.bin: %v", err)
	}
	// Read the protobuf encoded data back from the file
	protbufEncodedData, err := os.ReadFile("GameAPISession_protobuf.bin")
	if err != nil {
		t.Fatalf("failed to read GameAPISession_protobuf.bin: %v", err)
	}

	// Verify the size of the protobuf encoded data
	if len(protbufEncodedData) == 0 {
		t.Fatalf("no data was written to GameAPISession_protobuf.bin")
	}
	// Print the size of the protobuf encoded data
	t.Logf("Protobuf encoded data size: %d bytes", len(protbufEncodedData))
	// Verify the size of the original data
	if len(data) == 0 {
		t.Fatalf("no data was read from GameAPISession.json")
	}
	t.Logf("Original data size: %d bytes", len(data))
	// Print the compression ratio
	compressionRatio := float64(len(protbufEncodedData)) / float64(len(data))
	t.Logf("Compression ratio: %.2f", compressionRatio)
	// Check if the compression ratio is less than 0.5
	if compressionRatio >= 0.5 {
		t.Errorf("compression ratio is too high: %.2f, expected less than 0.5", compressionRatio)
	} else {
		t.Logf("compression ratio is acceptable: %.2f", compressionRatio)
	}
	// Check if the protobuf encoded data is smaller than the original data
	if len(protbufEncodedData) >= len(data) {
		t.Errorf("protobuf encoded data is not smaller than original data: %d >= %d", len(protbufEncodedData), len(data))
	} else {
		t.Logf("protobuf encoded data is smaller than original data: %d < %d", len(protbufEncodedData), len(data))
	}
}

func loadFixture(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	return data
}

func TestGameAPISessionEncoding(t *testing.T) {

	originalJSON := `{"disc":{"position":[-1.7870001,0.162,34.439003],"forward":[-0.047000002,-0.98400003,-0.171],"left":[-0.89300007,-0.035,0.44800001],"up":[-0.44700003,0.17300001,-0.87700003],"velocity":[3.1090002,3.4870002,-3.0960002],"bounce_count":0},"orange_team_restart_request":0,"sessionid":"892E23E2-8272-4B07-85E8-5E291AC85F93","game_clock_display":"10:00.00","game_status":"round_over","sessionip":"216.104.43.94","match_type":"Echo_Arena_Private","map_name":"mpl_arena_a","right_shoulder_pressed2":0.0,"teams":[{"players":[{"name":"Loveridge-","rhand":{"pos":[4.96,7.6360002,-8.3460007],"forward":[-0.42700002,-0.58500004,0.68900001],"left":[-0.64500004,0.73200005,0.22100002],"up":[-0.634,-0.35000002,-0.69000006]},"playerid":0,"stats":{"possession_time":140.79926,"points":13,"saves":2,"goals":0,"stuns":30,"passes":0,"catches":0,"steals":3,"blocks":0,"interceptions":0,"assists":2,"shots_taken":12},"userid":1583780395029509,"number":0,"level":50,"stunned":false,"ping":54,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[4.9720001,7.2960005,-8.9040003],"forward":[-0.54400003,-0.252,0.80100006],"left":[-0.83900005,0.164,-0.51800001],"up":[-0.001,-0.95400006,-0.30100003]},"body":{"position":[4.9720001,7.2960005,-8.9040003],"forward":[-0.26100001,-0.095000006,0.96100003],"left":[-0.96400005,0.083000004,-0.25400001],"up":[-0.056000002,-0.99200004,-0.11400001]},"holding_right":"none","lhand":{"pos":[5.1880002,7.9290004,-8.8570004],"forward":[0.87400001,0.29100001,-0.38900003],"left":[0.48500001,-0.56800002,0.66500002],"up":[-0.027000001,-0.77000004,-0.63800001]},"blocking":false,"velocity":[-2.5210001,-0.99000007,3.2860003]},{"name":"Hazzzah","rhand":{"pos":[-12.733001,-0.39200002,33.644001],"forward":[-0.68400002,-0.71400005,0.148],"left":[-0.296,0.086000003,-0.95100003],"up":[0.66700006,-0.69500005,-0.27000001]},"playerid":4,"stats":{"possession_time":81.235016,"points":11,"saves":3,"goals":0,"stuns":22,"passes":0,"catches":0,"steals":0,"blocks":0,"interceptions":0,"assists":4,"shots_taken":6},"userid":1403213729748087,"number":0,"level":50,"stunned":false,"ping":23,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-12.144001,0.37900001,33.359001],"forward":[0.91500002,-0.30200002,-0.26700002],"left":[-0.33700001,-0.21000001,-0.91800004],"up":[0.22100002,0.93000007,-0.294]},"body":{"position":[-12.144001,0.37900001,33.359001],"forward":[0.99700004,-0.001,0.076000005],"left":[0.076000005,-0.0020000001,-0.99700004],"up":[0.001,1.0,-0.0020000001]},"holding_right":"none","lhand":{"pos":[-12.225,-0.55900002,33.152],"forward":[0.20100001,-0.94600004,-0.25300002],"left":[0.60500002,0.32300001,-0.72800004],"up":[0.77000004,-0.0070000002,0.63800001]},"blocking":false,"velocity":[3.8710003,1.08,-0.36000001]},{"name":"WhiteDragon7","rhand":{"pos":[-2.9510002,-0.92000002,27.303001],"forward":[-0.12100001,0.98100007,0.15400001],"left":[0.26500002,-0.11800001,0.95700002],"up":[0.95700002,0.156,-0.24600001]},"playerid":5,"stats":{"possession_time":126.04878,"points":10,"saves":2,"goals":0,"stuns":45,"passes":0,"catches":0,"steals":1,"blocks":0,"interceptions":0,"assists":5,"shots_taken":8},"userid":1529343560460586,"number":0,"level":50,"stunned":false,"ping":62,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-2.4200001,-0.98100007,27.144001],"forward":[0.14300001,0.011000001,0.99000007],"left":[0.98500007,-0.096000001,-0.141],"up":[0.094000004,0.99500006,-0.025]},"body":{"position":[-2.4200001,-0.98100007,27.144001],"forward":[0.064000003,0.054000001,0.99700004],"left":[0.99300003,-0.104,-0.058000002],"up":[0.1,0.99300003,-0.060000002]},"holding_right":"none","lhand":{"pos":[-2.0630002,-1.404,27.432001],"forward":[-0.208,0.44600001,0.87000006],"left":[0.77600002,0.61700004,-0.13000001],"up":[-0.59500003,0.648,-0.47500002]},"blocking":false,"velocity":[3.9190001,0.40600002,3.062]},{"name":"Palidore","rhand":{"pos":[-0.97400004,0.10700001,35.621002],"forward":[0.87700003,0.15700001,0.45400003],"left":[0.47600001,-0.163,-0.86400002],"up":[-0.062000003,0.97400004,-0.21800001]},"playerid":9,"stats":{"possession_time":87.97332,"points":5,"saves":2,"goals":0,"stuns":37,"passes":0,"catches":0,"steals":0,"blocks":0,"interceptions":0,"assists":2,"shots_taken":4},"userid":1426294494120549,"number":0,"level":50,"stunned":false,"ping":88,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-1.143,0.46200001,35.623001],"forward":[0.41700003,-0.87000006,0.26100001],"left":[0.52000004,-0.0070000002,-0.85400003],"up":[0.74500006,0.49200001,0.45000002]},"body":{"position":[-1.143,0.46200001,35.623001],"forward":[0.82300001,0.0,0.56700003],"left":[0.56700003,-0.0020000001,-0.82300001],"up":[0.001,1.0,-0.0020000001]},"holding_right":"none","lhand":{"pos":[-1.1,0.079000004,35.448002],"forward":[0.80000001,-0.56300002,0.20700002],"left":[0.41100001,0.26300001,-0.87300003],"up":[0.43700001,0.78400004,0.44200003]},"blocking":false,"velocity":[2.8610001,-3.4710002,3.4620001]},{"name":"Cruisen","rhand":{"pos":[-10.527,0.60800004,-2.0900002],"forward":[0.87300003,-0.11400001,-0.47400004],"left":[-0.47600001,0.011000001,-0.87900007],"up":[0.105,0.99300003,-0.045000002]},"playerid":7,"stats":{"possession_time":159.86539,"points":3,"saves":5,"goals":0,"stuns":32,"passes":0,"catches":0,"steals":0,"blocks":0,"interceptions":0,"assists":1,"shots_taken":3},"userid":2455552844517585,"number":0,"level":50,"stunned":false,"ping":68,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-10.807,1.483,-1.8520001],"forward":[0.141,-0.34900001,0.92700005],"left":[0.99000007,0.052000001,-0.13100001],"up":[-0.0020000001,0.93600005,0.35300002]},"body":{"position":[-10.807,1.483,-1.8520001],"forward":[0.56200004,0.0,0.82700002],"left":[0.82700002,-0.0020000001,-0.56200004],"up":[0.0020000001,1.0,-0.001]},"holding_right":"none","lhand":{"pos":[-10.573001,0.73200005,-1.8140001],"forward":[0.27100003,-0.60600001,0.74800003],"left":[0.95500004,0.26500002,-0.132],"up":[-0.11800001,0.75000006,0.65000004]},"blocking":false,"velocity":[-1.279,0.33000001,-4.8710003]}],"team":"Team Splash","possession":false,"stats":{"points":42,"possession_time":595.92175,"interceptions":0,"blocks":0,"steals":4,"catches":0,"passes":0,"saves":14,"goals":0,"stuns":166,"assists":14,"shots_taken":33}},{"players":[{"name":"NtsFranz","rhand":{"pos":[-1.286,-0.93200004,33.559002],"forward":[-0.11100001,-0.73000002,-0.67400002],"left":[-0.099000007,-0.66700006,0.73900002],"up":[-0.98900002,0.149,0.0020000001]},"playerid":6,"stats":{"possession_time":150.77986,"points":6,"saves":2,"goals":0,"stuns":41,"passes":0,"catches":0,"steals":2,"blocks":0,"interceptions":0,"assists":2,"shots_taken":4},"userid":1644301615643409,"number":0,"level":50,"stunned":false,"ping":63,"packetlossratio":0.0,"invulnerable":false,"holding_left":"disc","possession":false,"head":{"position":[-1.4080001,-0.073000006,33.883003],"forward":[-0.098000005,-0.33800003,-0.93600005],"left":[-0.99300003,0.101,0.067000002],"up":[0.072000004,0.93600005,-0.34600002]},"body":{"position":[-1.4080001,-0.073000006,33.883003],"forward":[-0.62800002,0.0,-0.77900004],"left":[-0.77900004,0.0,0.62800002],"up":[0.0,1.0,0.0]},"holding_right":"none","lhand":{"pos":[-1.6670001,0.19800001,34.412003],"forward":[-0.083000004,0.98900002,0.126],"left":[0.84500003,0.003,0.53500003],"up":[0.528,0.15100001,-0.83600003]},"blocking":false,"velocity":[2.4950001,2.5930002,-1.7470001]},{"name":"speedy_v","rhand":{"pos":[-1.049,-0.52900004,34.280003],"forward":[-0.94300002,-0.23800001,0.23300001],"left":[0.32900003,-0.55000001,0.76800001],"up":[-0.055000003,0.80100006,0.597]},"playerid":1,"stats":{"possession_time":127.35746,"points":4,"saves":2,"goals":0,"stuns":43,"passes":0,"catches":0,"steals":2,"blocks":0,"interceptions":0,"assists":4,"shots_taken":6},"userid":1765442746814333,"number":0,"level":50,"stunned":false,"ping":51,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":true,"head":{"position":[-0.97400004,0.042000003,34.518002],"forward":[-0.89500004,-0.44300002,-0.054000001],"left":[-0.081,0.041000001,0.99600005],"up":[-0.43900001,0.89500004,-0.072000004]},"body":{"position":[-0.97400004,0.042000003,34.518002],"forward":[-0.99800003,0.0,0.066],"left":[0.066,0.0020000001,0.99800003],"up":[0.0,1.0,-0.0020000001]},"holding_right":"none","lhand":{"pos":[-1.373,-0.40400001,34.952003],"forward":[-0.59000003,0.2,0.78200006],"left":[0.75800002,0.47000003,0.45200002],"up":[-0.27700001,0.86000001,-0.42900002]},"blocking":false,"velocity":[-2.6100001,-0.27000001,-3.5990002]},{"name":"Dual-","rhand":{"pos":[-5.6790004,1.072,31.664001],"forward":[0.65300006,-0.73600006,0.178],"left":[-0.156,-0.36000001,-0.92000002],"up":[0.74100006,0.57300001,-0.35000002]},"playerid":8,"stats":{"possession_time":76.696434,"points":2,"saves":2,"goals":0,"stuns":30,"passes":0,"catches":0,"steals":0,"blocks":0,"interceptions":0,"assists":2,"shots_taken":8},"userid":1589255294473366,"number":0,"level":50,"stunned":false,"ping":43,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-5.5880003,1.8720001,31.496002],"forward":[0.91100007,-0.40500003,0.074000001],"left":[0.089000002,0.017000001,-0.99600005],"up":[0.40200001,0.91400003,0.051000003]},"body":{"position":[-5.5880003,1.8720001,31.496002],"forward":[0.99900007,-0.001,0.045000002],"left":[0.045000002,-0.001,-0.99900007],"up":[0.0020000001,1.0,-0.001]},"holding_right":"none","lhand":{"pos":[-5.4660001,1.057,31.354002],"forward":[0.75700003,-0.61400002,0.22500001],"left":[0.63300002,0.60100001,-0.48800004],"up":[0.16500001,0.51100004,0.84300005]},"blocking":false,"velocity":[-0.30000001,0.038000003,0.073000006]},{"name":"qlyoung","rhand":{"pos":[5.7580004,0.92900002,-9.8730001],"forward":[-0.16000001,-0.94200003,0.29500002],"left":[0.98700005,-0.15900001,0.028000001],"up":[0.020000001,0.29500002,0.95500004]},"playerid":3,"stats":{"possession_time":106.62419,"points":0,"saves":1,"goals":0,"stuns":35,"passes":0,"catches":0,"steals":0,"blocks":0,"interceptions":0,"assists":0,"shots_taken":4},"userid":1956271971053349,"number":0,"level":50,"stunned":false,"ping":53,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[5.9960003,1.835,-9.9510002],"forward":[-0.24700001,0.17500001,0.95300007],"left":[0.96900004,0.053000003,0.24200001],"up":[-0.0080000004,0.98300004,-0.18300001]},"body":{"position":[5.9960003,1.835,-9.9510002],"forward":[0.27700001,0.0020000001,0.96100003],"left":[0.96100003,-0.001,-0.27700001],"up":[0.0,1.0,-0.0020000001]},"holding_right":"none","lhand":{"pos":[6.2090001,0.89800006,-10.21],"forward":[0.071000002,-0.98800004,0.14],"left":[0.96500003,0.033,-0.259],"up":[0.25100002,0.15400001,0.95600003]},"blocking":false,"velocity":[-2.5650001,1.2600001,2.924]},{"name":"SputnikKobra","rhand":{"pos":[-10.548,6.2750001,30.190001],"forward":[-0.003,-0.73200005,0.68200004],"left":[-0.98700005,-0.108,-0.12100001],"up":[0.162,-0.67300004,-0.72200006]},"playerid":2,"stats":{"possession_time":100.44829,"points":6,"saves":4,"goals":0,"stuns":20,"passes":0,"catches":0,"steals":1,"blocks":0,"interceptions":0,"assists":0,"shots_taken":7},"userid":1865696153442253,"number":0,"level":50,"stunned":false,"ping":42,"packetlossratio":0.0,"invulnerable":false,"holding_left":"none","possession":false,"head":{"position":[-10.683001,7.1520004,29.917002],"forward":[0.86000001,-0.38700002,0.33100003],"left":[0.264,-0.21700001,-0.94000006],"up":[0.43600002,0.89600003,-0.084000006]},"body":{"position":[-10.683001,7.1520004,29.917002],"forward":[0.98200005,0.084000006,0.171],"left":[0.16600001,0.066,-0.98400003],"up":[-0.094000004,0.99400002,0.051000003]},"holding_right":"none","lhand":{"pos":[-10.698001,7.0980005,29.617001],"forward":[-0.043000001,-0.97500002,-0.21700001],"left":[0.70500004,-0.18300001,0.68500006],"up":[-0.708,-0.123,0.69500005]},"blocking":false,"velocity":[0.40600002,-0.045000002,0.40600002]}],"team":"ORANGE TEAM","possession":true,"stats":{"points":18,"possession_time":561.90625,"interceptions":0,"blocks":0,"steals":5,"catches":0,"passes":0,"saves":11,"goals":0,"stuns":169,"assists":8,"shots_taken":29}},{"team":"SPECTATORS","possession":false}],"blue_round_score":3,"orange_points":8,"player":{"vr_left":[0.91500002,-0.21900001,-0.33700001],"vr_position":[0.19700001,-0.018000001,0.35700002],"vr_forward":[0.39000002,0.27500001,0.87900007],"vr_up":[0.1,0.93600005,-0.33700001]},"private_match":true,"blue_team_restart_request":0,"tournament_match":false,"orange_round_score":0,"total_round_count":3,"left_shoulder_pressed2":0.0,"left_shoulder_pressed":1.0,"pause":{"paused_state":"unpaused","unpaused_team":"none","paused_requested_team":"none","unpaused_timer":0.0,"paused_timer":0.0},"right_shoulder_pressed":0.0,"blue_points":15,"last_throw":{"arm_speed":0.35953924,"total_speed":4.305512,"off_axis_spin_deg":-23.038441,"wrist_throw_penalty":0.10544699,"rot_per_sec":0.12869783,"pot_speed_from_rot":0.09703587,"speed_from_arm":0.35953927,"speed_from_movement":3.9750874,"speed_from_wrist":-0.029114723,"wrist_align_to_throw_deg":102.79651,"throw_align_to_movement_deg":23.276352,"off_axis_penalty":0.0074819326,"throw_move_penalty":0.026893407},"client_name":"NtsFranz","game_clock":600.0,"possession":[1,1],"last_score":{"disc_speed":9.3109922,"team":"orange","goal_type":"INSIDE SHOT","point_amount":2,"distance_thrown":0.48852247,"person_scored":"NtsFranz","assist_scored":"speedy_v"},"err_code":0}`

	session := &gameapi.SessionResponse{
		Disc: &gameapi.Disc{
			Position:    []float64{-1.7870001, 0.162, 34.439003},
			Forward:     []float64{-0.047000002, -0.98400003, -0.171},
			Left:        []float64{-0.89300007, -0.035, 0.44800001},
			Up:          []float64{-0.44700003, 0.17300001, -0.87700003},
			Velocity:    []float64{3.1090002, 3.4870002, -3.0960002},
			BounceCount: 0,
		},
		OrangeTeamRestartRequest: 0,
		SessionID:                "892E23E2-8272-4B07-85E8-5E291AC85F93",
		GameClockDisplay:         "10:00.00",
		GameStatus:               "round_over",
		SessionIP:                "216.104.43.94",
		MatchType:                "Echo_Arena_Private",
		MapName:                  "mpl_arena_a",
		RightShoulderPressed2:    0.0,
		Teams: []*gameapi.Team{
			{
				TeamName:      "Team Splash",
				HasPossession: false,
				Stats: &gameapi.TeamStats{
					Points:         42,
					PossessionTime: 595.92175,
					Interceptions:  0,
					Blocks:         0,
					Steals:         4,
					Catches:        0,
					Passes:         0,
					Saves:          14,
					Goals:          0,
					Stuns:          166,
					Assists:        14,
					ShotsTaken:     33,
				},
				Players: []*gameapi.TeamMember{
					{
						DisplayName:  "Loveridge-",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 140.79926,
							Points:         13,
							Saves:          2,
							Goals:          0,
							Stuns:          30,
							Passes:         0,
							Catches:        0,
							Steals:         3,
							Blocks:         0,
							Interceptions:  0,
							Assists:        2,
							ShotsTaken:     12,
						},
						AccountNumber:    1583780395029509,
						SlotNumber:       0,
						Level:            50,
						IsStunned:        false,
						Ping:             54,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{-2.5210001, -0.99000007, 3.2860003},
						Head: &gameapi.BodyPart{
							Position: []float64{4.9720001, 7.2960005, -8.9040003},
							Forward:  []float64{-0.54400003, -0.252, 0.80100006},
							Left:     []float64{-0.83900005, 0.164, -0.51800001},
							Up:       []float64{-0.001, -0.95400006, -0.30100003},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{4.9720001, 7.2960005, -8.9040003},
							Forward:  []float64{-0.26100001, -0.095000006, 0.96100003},
							Left:     []float64{-0.96400005, 0.083000004, -0.25400001},
							Up:       []float64{-0.056000002, -0.99200004, -0.11400001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{5.1880002, 7.9290004, -8.8570004},
							Forward: []float64{0.87400001, 0.29100001, -0.38900003},
							Left:    []float64{0.48500001, -0.56800002, 0.66500002},
							Up:      []float64{-0.027000001, -0.77000004, -0.63800001},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{4.96, 7.6360002, -8.3460007},
							Forward: []float64{-0.42700002, -0.58500004, 0.68900001},
							Left:    []float64{-0.64500004, 0.73200005, 0.22100002},
							Up:      []float64{-0.634, -0.35000002, -0.69000006},
						},
					},
					{
						DisplayName: "Hazzzah",
						SlotNumber:  4,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 81.235016,
							Points:         11,
							Saves:          3,
							Goals:          0,
							Stuns:          22,
							Passes:         0,
							Catches:        0,
							Steals:         0,
							Blocks:         0,
							Interceptions:  0,
							Assists:        4,
							ShotsTaken:     6,
						},
						AccountNumber:    1403213729748087,
						JerseyNumber:     0,
						Level:            50,
						IsStunned:        false,
						Ping:             23,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{3.8710003, 1.08, -0.36000001},
						Head: &gameapi.BodyPart{
							Position: []float64{-12.144001, 0.37900001, 33.359001},
							Forward:  []float64{0.91500002, -0.30200002, -0.26700002},
							Left:     []float64{-0.33700001, -0.21000001, -0.91800004},
							Up:       []float64{0.22100002, 0.93000007, -0.294},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-12.144001, 0.37900001, 33.359001},
							Forward:  []float64{0.99700004, -0.001, 0.076000005},
							Left:     []float64{0.076000005, -0.0020000001, -0.99700004},
							Up:       []float64{0.001, 1, -0.0020000001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-12.225, -0.55900002, 33.152},
							Forward: []float64{0.20100001, -0.94600004, -0.25300002},
							Left:    []float64{0.60500002, 0.32300001, -0.72800004},
							Up:      []float64{0.77000004, -0.0070000002, 0.63800001},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-12.733001, -0.39200002, 33.644001},
							Forward: []float64{-0.68400002, -0.71400005, 0.148},
							Left:    []float64{-0.296, 0.086000003, -0.95100003},
							Up:      []float64{0.66700006, -0.69500005, -0.27000001},
						},
					},
					{
						DisplayName: "WhiteDragon7",
						SlotNumber:  5,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 126.04878,
							Points:         10,
							Saves:          2,
							Goals:          0,
							Stuns:          45,
							Passes:         0,
							Catches:        0,
							Steals:         1,
							Blocks:         0,
							Interceptions:  0,
							Assists:        5,
							ShotsTaken:     8,
						},
						AccountNumber:    1529343560460586,
						JerseyNumber:     0,
						Level:            50,
						IsStunned:        false,
						Ping:             62,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{3.9190001, 0.40600002, 3.062},
						Head: &gameapi.BodyPart{
							Position: []float64{-2.4200001, -0.98100007, 27.144001},
							Forward:  []float64{0.14300001, 0.011000001, 0.99000007},
							Left:     []float64{0.98500007, -0.096000001, -0.141},
							Up:       []float64{0.094000004, 0.99500006, -0.025},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-2.4200001, -0.98100007, 27.144001},
							Forward:  []float64{0.064000003, 0.054000001, 0.99700004},
							Left:     []float64{0.99300003, -0.104, -0.058000002},
							Up:       []float64{0.1, 0.99300003, -0.060000002},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-2.0630002, -1.404, 27.432001},
							Forward: []float64{-0.208, 0.44600001, 0.87000006},
							Left:    []float64{0.77600002, 0.61700004, -0.13000001},
							Up:      []float64{-0.59500003, 0.648, -0.47500002},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-2.9510002, -0.92000002, 27.303001},
							Forward: []float64{-0.12100001, 0.98100007, 0.15400001},
							Left:    []float64{0.26500002, -0.11800001, 0.95700002},
							Up:      []float64{0.95700002, 0.156, -0.24600001},
						},
					},
					{
						DisplayName: "Palidore",
						SlotNumber:  9,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 87.97332,
							Points:         5,
							Saves:          2,
							Goals:          0,
							Stuns:          37,
							Passes:         0,
							Catches:        0,
							Steals:         0,
							Blocks:         0,
							Interceptions:  0,
							Assists:        2,
							ShotsTaken:     4,
						},
						AccountNumber:    1426294494120549,
						JerseyNumber:     0,
						Level:            50,
						IsStunned:        false,
						Ping:             88,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{2.8610001, -3.4710002, 3.4620001},
						Head: &gameapi.BodyPart{
							Position: []float64{-1.143, 0.46200001, 35.623001},
							Forward:  []float64{0.41700003, -0.87000006, 0.26100001},
							Left:     []float64{0.52000004, -0.0070000002, -0.85400003},
							Up:       []float64{0.74500006, 0.49200001, 0.45000002},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-1.143, 0.46200001, 35.623001},
							Forward:  []float64{0.82300001, 0.0, 0.56700003},
							Left:     []float64{0.56700003, -0.0020000001, -0.82300001},
							Up:       []float64{0.001, 1.0, -0.0020000001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-1.1, 0.079000004, 35.448002},
							Forward: []float64{0.80000001, -0.56300002, 0.20700002},
							Left:    []float64{0.41100001, 0.26300001, -0.87300003},
							Up:      []float64{0.43700001, 0.78400004, 0.44200003},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-0.97400004, 0.10700001, 35.621002},
							Forward: []float64{0.87700003, 0.15700001, 0.45400003},
							Left:    []float64{0.47600001, -0.163, -0.86400002},
							Up:      []float64{-0.062000003, 0.97400004, -0.21800001},
						},
					},
					{
						DisplayName: "Cruisen",
						SlotNumber:  7,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 159.86539,
							Points:         3,
							Saves:          5,
							Goals:          0,
							Stuns:          32,
							Passes:         0,
							Catches:        0,
							Steals:         0,
							Blocks:         0,
							Interceptions:  0,
							Assists:        1,
							ShotsTaken:     3,
						},
						AccountNumber:    2455552844517585,
						JerseyNumber:     0,
						Level:            50,
						IsStunned:        false,
						Ping:             68,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{-1.279, 0.33000001, -4.8710003},
						Head: &gameapi.BodyPart{
							Position: []float64{-10.807, 1.483, -1.8520001},
							Forward:  []float64{0.141, -0.34900001, 0.92700005},
							Left:     []float64{0.99000007, 0.052000001, -0.13100001},
							Up:       []float64{-0.0020000001, 0.93600005, 0.35300002},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-10.807, 1.483, -1.8520001},
							Forward:  []float64{0.56200004, 0.0, 0.82700002},
							Left:     []float64{0.82700002, -0.0020000001, -0.56200004},
							Up:       []float64{0.0020000001, 1.0, -0.001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-10.573001, 0.73200005, -1.8140001},
							Forward: []float64{0.27100003, -0.60600001, 0.74800003},
							Left:    []float64{0.95500004, 0.26500002, -0.132},
							Up:      []float64{-0.11800001, 0.75000006, 0.65000004},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-10.527, 0.60800004, -2.0900002},
							Forward: []float64{0.87300003, -0.11400001, -0.47400004},
							Left:    []float64{-0.47600001, 0.011000001, -0.87900007},
							Up:      []float64{0.105, 0.99300003, -0.045000002},
						},
					},
				},
			},
			{
				Players: []*gameapi.TeamMember{
					{
						DisplayName:  "NtsFranz",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 150.77986,
							Points:         6,
							Saves:          2,
							Goals:          0,
							Stuns:          41,
							Passes:         0,
							Catches:        0,
							Steals:         2,
							Blocks:         0,
							Interceptions:  0,
							Assists:        2,
							ShotsTaken:     4,
						},
						AccountNumber:    1644301615643409,
						SlotNumber:       6,
						Level:            50,
						IsStunned:        false,
						Ping:             63,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "disc",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{2.4950001, 2.5930002, -1.7470001},
						Head: &gameapi.BodyPart{
							Position: []float64{-1.4080001, -0.073000006, 33.883003},
							Forward:  []float64{-0.098000005, -0.33800003, -0.93600005},
							Left:     []float64{-0.99300003, 0.101, 0.067000002},
							Up:       []float64{0.072000004, 0.93600005, -0.34600002},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-1.4080001, -0.073000006, 33.883003},
							Forward:  []float64{-0.62800002, 0.0, -0.77900004},
							Left:     []float64{-0.77900004, 0.0, 0.62800002},
							Up:       []float64{0.0, 1.0, 0.0},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-1.6670001, 0.19800001, 34.412003},
							Forward: []float64{-0.083000004, 0.98900002, 0.126},
							Left:    []float64{0.84500003, 0.003, 0.53500003},
							Up:      []float64{0.528, 0.15100001, -0.83600003},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-1.286, -0.93200004, 33.559002},
							Forward: []float64{-0.11100001, -0.73000002, -0.67400002},
							Left:    []float64{-0.099000007, -0.66700006, 0.73900002},
							Up:      []float64{-0.98900002, 0.149, 0.0020000001},
						},
					},
					{
						DisplayName:  "speedy_v",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 127.35746,
							Points:         4,
							Saves:          2,
							Goals:          0,
							Stuns:          43,
							Passes:         0,
							Catches:        0,
							Steals:         2,
							Blocks:         0,
							Interceptions:  0,
							Assists:        4,
							ShotsTaken:     6,
						},
						AccountNumber:    1765442746814333,
						SlotNumber:       1,
						Level:            50,
						IsStunned:        false,
						Ping:             51,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    true,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{-2.6100001, -0.27000001, -3.5990002},
						Head: &gameapi.BodyPart{
							Position: []float64{-0.97400004, 0.042000003, 34.518002},
							Forward:  []float64{-0.89500004, -0.44300002, -0.054000001},
							Left:     []float64{-0.081, 0.041000001, 0.99600005},
							Up:       []float64{-0.43900001, 0.89500004, -0.072000004},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-0.97400004, 0.042000003, 34.518002},
							Forward:  []float64{-0.99800003, 0.0, 0.066},
							Left:     []float64{0.066, 0.0020000001, 0.99800003},
							Up:       []float64{0.0, 1.0, -0.0020000001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-1.373, -0.40400001, 34.952003},
							Forward: []float64{-0.59000003, 0.2, 0.78200006},
							Left:    []float64{0.75800002, 0.47000003, 0.45200002},
							Up:      []float64{-0.27700001, 0.86000001, -0.42900002},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-1.049, -0.52900004, 34.280003},
							Forward: []float64{-0.94300002, -0.23800001, 0.23300001},
							Left:    []float64{0.32900003, -0.55000001, 0.76800001},
							Up:      []float64{-0.055000003, 0.80100006, 0.597},
						},
					},
					{
						DisplayName:  "Dual-",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 76.696434,
							Points:         2,
							Saves:          2,
							Goals:          0,
							Stuns:          30,
							Passes:         0,
							Catches:        0,
							Steals:         0,
							Blocks:         0,
							Interceptions:  0,
							Assists:        2,
							ShotsTaken:     8,
						},
						AccountNumber:    1589255294473366,
						SlotNumber:       8,
						Level:            50,
						IsStunned:        false,
						Ping:             43,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{-0.30000001, 0.038000003, 0.073000006},
						Head: &gameapi.BodyPart{
							Position: []float64{-5.5880003, 1.8720001, 31.496002},
							Forward:  []float64{0.91100007, -0.40500003, 0.074000001},
							Left:     []float64{0.089000002, 0.017000001, -0.99600005},
							Up:       []float64{0.40200001, 0.91400003, 0.051000003},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-5.5880003, 1.8720001, 31.496002},
							Forward:  []float64{0.99900007, -0.001, 0.045000002},
							Left:     []float64{0.045000002, -0.001, -0.99900007},
							Up:       []float64{0.0020000001, 1.0, -0.001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-5.4660001, 1.057, 31.354002},
							Forward: []float64{0.75700003, -0.61400002, 0.22500001},
							Left:    []float64{0.63300002, 0.60100001, -0.48800004},
							Up:      []float64{0.16500001, 0.51100004, 0.84300005},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-5.6790004, 1.072, 31.664001},
							Forward: []float64{0.65300006, -0.73600006, 0.178},
							Left:    []float64{-0.156, -0.36000001, -0.92000002},
							Up:      []float64{0.74100006, 0.57300001, -0.35000002},
						},
					},
					{
						DisplayName:  "qlyoung",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 106.62419,
							Points:         0,
							Saves:          1,
							Goals:          0,
							Stuns:          35,
							Passes:         0,
							Catches:        0,
							Steals:         0,
							Blocks:         0,
							Interceptions:  0,
							Assists:        0,
							ShotsTaken:     4,
						},
						AccountNumber:    1956271971053349,
						SlotNumber:       3,
						Level:            50,
						IsStunned:        false,
						Ping:             53,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{-2.5650001, 1.2600001, 2.924},
						Head: &gameapi.BodyPart{
							Position: []float64{5.9960003, 1.835, -9.9510002},
							Forward:  []float64{-0.24700001, 0.17500001, 0.95300007},
							Left:     []float64{0.96900004, 0.053000003, 0.24200001},
							Up:       []float64{-0.0080000004, 0.98300004, -0.18300001},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{5.9960003, 1.835, -9.9510002},
							Forward:  []float64{0.27700001, 0.0020000001, 0.96100003},
							Left:     []float64{0.96100003, -0.001, -0.27700001},
							Up:       []float64{0.0, 1.0, -0.0020000001},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{6.2090001, 0.89800006, -10.21},
							Forward: []float64{0.071000002, -0.98800004, 0.14},
							Left:    []float64{0.96500003, 0.033, -0.259},
							Up:      []float64{0.25100002, 0.15400001, 0.95600003},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{5.7580004, 0.92900002, -9.8730001},
							Forward: []float64{-0.16000001, -0.94200003, 0.29500002},
							Left:    []float64{0.98700005, -0.15900001, 0.028000001},
							Up:      []float64{0.020000001, 0.29500002, 0.95500004},
						},
					},
					{
						DisplayName:  "SputnikKobra",
						JerseyNumber: 0,
						Stats: &gameapi.PlayerStats{
							PossessionTime: 100.44829,
							Points:         6,
							Saves:          4,
							Goals:          0,
							Stuns:          20,
							Passes:         0,
							Catches:        0,
							Steals:         1,
							Blocks:         0,
							Interceptions:  0,
							Assists:        0,
							ShotsTaken:     7,
						},
						AccountNumber:    1865696153442253,
						SlotNumber:       2,
						Level:            50,
						IsStunned:        false,
						Ping:             42,
						PacketLossRatio:  0.0,
						IsInvulnerable:   false,
						LeftHoldingOnto:  "none",
						HasPossession:    false,
						RightHoldingOnto: "none",
						IsBlocking:       false,
						Velocity:         []float64{0.40600002, -0.045000002, 0.40600002},
						Head: &gameapi.BodyPart{
							Position: []float64{-10.683001, 7.1520004, 29.917002},
							Forward:  []float64{0.86000001, -0.38700002, 0.33100003},
							Left:     []float64{0.264, -0.21700001, -0.94000006},
							Up:       []float64{0.43600002, 0.89600003, -0.084000006},
						},
						Body: &gameapi.BodyPart{
							Position: []float64{-10.683001, 7.1520004, 29.917002},
							Forward:  []float64{0.98200005, 0.084000006, 0.171},
							Left:     []float64{0.16600001, 0.066, -0.98400003},
							Up:       []float64{-0.094000004, 0.99400002, 0.051000003},
						},
						LeftHand: &gameapi.BodyPart{
							Pos:     []float64{-10.698001, 7.0980005, 29.617001},
							Forward: []float64{-0.043000001, -0.97500002, -0.21700001},
							Left:    []float64{0.70500004, -0.18300001, 0.68500006},
							Up:      []float64{-0.708, -0.123, 0.69500005},
						},
						RightHand: &gameapi.BodyPart{
							Pos:     []float64{-10.548, 6.2750001, 30.190001},
							Forward: []float64{-0.003, -0.73200005, 0.68200004},
							Left:    []float64{-0.98700005, -0.108, -0.12100001},
							Up:      []float64{0.162, -0.67300004, -0.72200006},
						},
					},
				},

				TeamName:      "ORANGE TEAM",
				HasPossession: true,
				Stats: &gameapi.TeamStats{
					Points:         18,
					PossessionTime: 561.90625,
					Interceptions:  0,
					Blocks:         0,
					Steals:         5,
					Catches:        0,
					Passes:         0,
					Saves:          11,
					Goals:          0,
					Stuns:          169,
					Assists:        8,
					ShotsTaken:     29,
				},
			},
			{
				TeamName:      "SPECTATORS",
				HasPossession: false,
			},
		},
		BlueRoundScore: 3,
		OrangePoints:   8,
		Player: &gameapi.PlayerRoot{
			VrLeft:     []float64{0.91500002, -0.21900001, -0.33700001},
			VrPosition: []float64{0.19700001, -0.018000001, 0.35700002},
			VrForward:  []float64{0.39000002, 0.27500001, 0.87900007},
			VrUp:       []float64{0.1, 0.93600005, -0.33700001},
		},
		PrivateMatch:           true,
		BlueTeamRestartRequest: 0,
		TournamentMatch:        false,
		OrangeRoundScore:       0,
		TotalRoundCount:        3,
		LeftShoulderPressed2:   0.0,
		LeftShoulderPressed:    1.0,
		Pause: &gameapi.PauseState{
			PausedState:         "unpaused",
			UnpausedTeam:        "none",
			PausedRequestedTeam: "none",
			UnpausedTimer:       0.0,
			PausedTimer:         0.0,
		},
		RightShoulderPressed: 0.0,
		BluePoints:           15,
		LastThrow: &gameapi.LastThrowInfo{
			ArmSpeed:                0.35953924,
			TotalSpeed:              4.305512,
			OffAxisSpinDeg:          -23.038441,
			WristThrowPenalty:       0.10544699,
			RotPerSec:               0.12869783,
			PotSpeedFromRot:         0.09703587,
			SpeedFromArm:            0.35953927,
			SpeedFromMovement:       3.9750874,
			SpeedFromWrist:          -0.029114723,
			WristAlignToThrowDeg:    102.79651,
			ThrowAlignToMovementDeg: 23.276352,
			OffAxisPenalty:          0.0074819326,
			ThrowMovePenalty:        0.026893407,
		},
		ClientName: "NtsFranz",
		GameClock:  600.0,
		Possession: []int32{1, 1},
		LastScore: &gameapi.LastScore{
			DiscSpeed:      9.3109922,
			Team:           "orange",
			GoalType:       "INSIDE SHOT",
			PointAmount:    2,
			DistanceThrown: 0.48852247,
			PersonScored:   "NtsFranz",
			AssistScored:   "speedy_v",
		},
		ErrCode: 0,
	}

	// Remarshal the fixture JSON to ensure it matches the expected structure
	m := map[string]any{}
	if err := json.Unmarshal([]byte(originalJSON), &m); err != nil {
		t.Fatalf("failed to unmarshal fixture JSON: %v", err)
	}

	want, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal fixture JSON: %v", err)
	}

	// Marshall the session to JSON
	data, err := protojsonMarshaler.Marshal(session)
	if err != nil {
		t.Fatalf("failed to marshal session: %v", err)
	}

	m = map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal session JSON: %v", err)
	}

	// Remarshal the session JSON to ensure it matches the expected structure
	got, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal session JSON: %v", err)
	}

	// Write it to want.json and got.json
	if err := os.WriteFile("want.json", want, 0644); err != nil {
		t.Fatalf("failed to write want.json: %v", err)
	}
	if err := os.WriteFile("got.json", got, 0644); err != nil {
		t.Fatalf("failed to write got.json: %v", err)
	}

	// Compare the original fixture JSON with the marshaled session JSON
	if s := cmp.Diff(string(want), string(got)); s != "" {
		t.Errorf("session JSON mismatch (-want +got):\n%s", s)
	}
}
