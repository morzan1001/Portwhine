package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/v1/portwhinev1connect"
)

// CLI configuration
var (
	serverAddr  = envOrDefault("PWCTL_SERVER", "http://localhost:50051")
	accessToken = ""
	tokenFile   = ""
)

func main() {
	// Determine token file location
	home, _ := os.UserHomeDir()
	if home != "" {
		tokenFile = home + "/.pwctl_token"
	}

	// Load saved token
	if tokenFile != "" {
		if data, err := os.ReadFile(tokenFile); err == nil {
			accessToken = strings.TrimSpace(string(data))
		}
	}

	// Parse global flags (--server/-s) before the subcommand.
	rawArgs := os.Args[1:]
	var filteredArgs []string
	for i := 0; i < len(rawArgs); i++ {
		switch {
		case rawArgs[i] == "--server" || rawArgs[i] == "-s":
			if i+1 < len(rawArgs) {
				serverAddr = rawArgs[i+1]
				i++
			}
		case strings.HasPrefix(rawArgs[i], "--server="):
			serverAddr = strings.TrimPrefix(rawArgs[i], "--server=")
		case strings.HasPrefix(rawArgs[i], "-s="):
			serverAddr = strings.TrimPrefix(rawArgs[i], "-s=")
		default:
			filteredArgs = append(filteredArgs, rawArgs[i])
		}
	}

	if len(filteredArgs) < 1 {
		printUsage()
		os.Exit(1)
	}

	cmd := filteredArgs[0]
	args := filteredArgs[1:]

	var err error
	switch cmd {
	case "login":
		err = cmdLogin(args)
	case "register":
		err = cmdRegister(args)
	case "pipeline":
		err = dispatchPipeline(args)
	case "run":
		err = dispatchRun(args)
	case "worker":
		err = dispatchWorker(args)
	case "user":
		err = dispatchUser(args)
	case "apikey":
		err = dispatchAPIKey(args)
	case "team":
		err = dispatchTeam(args)
	case "permission":
		err = dispatchPermission(args)
	case "role":
		err = dispatchRole(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `pwctl - Portwhine CLI

Usage: pwctl <command> [subcommand] [args...]

Commands:
  login <username> <password>       Authenticate and save token
  register <username> <email> <pw>  Create a new user account

  pipeline list                     List all pipelines
  pipeline get <id>                 Get pipeline details
  pipeline create <json-file>       Create pipeline from JSON definition
  pipeline delete <id>              Delete a pipeline
  pipeline start <id>               Start a pipeline run
  pipeline stop <run-id>            Stop a pipeline run

  run list <pipeline-id>            List runs for a pipeline
  run status <run-id>               Get run status with per-node details
  run stream <run-id>               Stream live results
  run logs <run-id> <node-id>       Stream container logs [--tail N] [--follow]

  worker list                       List registered worker images
  worker register <name> <image>    Register a worker image
  worker delete <id>                Delete a worker image

  user list                         List users
  user get <id>                     Get user details
  user delete <id>                  Delete a user

  apikey list                       List API keys
  apikey create <name>              Create an API key
  apikey revoke <id>                Revoke an API key

  team list                         List teams
  team get <id>                     Get team details
  team create <name> [desc]         Create a team
  team delete <id>                  Delete a team
  team members <team-id>            List team members
  team add-member <team> <user> [role]  Add member to team
  team remove-member <team> <user>  Remove member from team
  team update-role <team> <user> <role> Update member role

  permission grant <flags>          Grant a permission
  permission revoke <id>            Revoke a permission
  permission list [flags]           List permissions
  permission my                     List my permissions

  role list                         List roles
  role create <name> [desc]         Create a custom role
  role update <id> <name> [desc]    Update a role
  role delete <id>                  Delete a role

Environment:
  PWCTL_SERVER  Operator address (default: http://localhost:50051)`)
}

func dispatchPipeline(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl pipeline <list|get|create|delete|start|stop>")
	}
	switch args[0] {
	case "list":
		return cmdPipelineList()
	case "get":
		return cmdPipelineGet(args[1:])
	case "create":
		return cmdPipelineCreate(args[1:])
	case "delete":
		return cmdPipelineDelete(args[1:])
	case "start":
		return cmdPipelineStart(args[1:])
	case "stop":
		return cmdPipelineStop(args[1:])
	default:
		return fmt.Errorf("unknown pipeline subcommand: %s", args[0])
	}
}

func dispatchRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl run <list|status|stream|logs>")
	}
	switch args[0] {
	case "list":
		return cmdRunList(args[1:])
	case "status":
		return cmdRunStatus(args[1:])
	case "stream":
		return cmdRunStream(args[1:])
	case "logs":
		return cmdRunLogs(args[1:])
	default:
		return fmt.Errorf("unknown run subcommand: %s", args[0])
	}
}

func dispatchWorker(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl worker <list|register|delete>")
	}
	switch args[0] {
	case "list":
		return cmdWorkerList()
	case "register":
		return cmdWorkerRegister(args[1:])
	case "delete":
		return cmdWorkerDelete(args[1:])
	default:
		return fmt.Errorf("unknown worker subcommand: %s", args[0])
	}
}

func dispatchUser(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl user <list|get|delete>")
	}
	switch args[0] {
	case "list":
		return cmdUserList()
	case "get":
		return cmdUserGet(args[1:])
	case "delete":
		return cmdUserDelete(args[1:])
	default:
		return fmt.Errorf("unknown user subcommand: %s", args[0])
	}
}

func dispatchAPIKey(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl apikey <list|create|revoke>")
	}
	switch args[0] {
	case "list":
		return cmdAPIKeyList()
	case "create":
		return cmdAPIKeyCreate(args[1:])
	case "revoke":
		return cmdAPIKeyRevoke(args[1:])
	default:
		return fmt.Errorf("unknown apikey subcommand: %s", args[0])
	}
}

// newClient creates an authenticated ConnectRPC client.
func newClient() portwhinev1connect.OperatorServiceClient {
	return portwhinev1connect.NewOperatorServiceClient(
		&http.Client{Timeout: 30 * time.Second},
		serverAddr,
	)
}

// authHeader returns a connect.Request with the auth header set.
func withAuth[T any](msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if accessToken != "" {
		req.Header().Set("Authorization", "Bearer "+accessToken)
	}
	return req
}

func saveToken(token string) {
	accessToken = token
	if tokenFile != "" {
		_ = os.WriteFile(tokenFile, []byte(token), 0600)
	}
}

// --- Auth commands ---

func cmdLogin(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl login <username> <password>")
	}

	client := newClient()
	resp, err := client.Login(context.Background(), connect.NewRequest(&portwhinev1.LoginRequest{
		Username: args[0],
		Password: args[1],
	}))
	if err != nil {
		return err
	}

	saveToken(resp.Msg.GetAccessToken())
	fmt.Println("Login successful. Token saved.")
	fmt.Printf("Expires at: %s\n", resp.Msg.GetExpiresAt().AsTime().Format(time.RFC3339))
	return nil
}

func cmdRegister(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: pwctl register <username> <email> <password>")
	}

	role := "user"
	if len(args) >= 4 {
		role = args[3]
	}

	client := newClient()
	resp, err := client.CreateUser(context.Background(), connect.NewRequest(&portwhinev1.CreateUserRequest{
		Username: args[0],
		Email:    args[1],
		Password: args[2],
		Role:     role,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("User created: %s\n", resp.Msg.GetUserId())
	return nil
}

// --- Pipeline commands ---

func cmdPipelineList() error {
	client := newClient()
	resp, err := client.ListPipelines(context.Background(), withAuth(&portwhinev1.ListPipelinesRequest{
		PageSize: 100,
	}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetPipelines()) == 0 {
		fmt.Println("No pipelines found.")
		return nil
	}

	fmt.Printf("%-36s  %-30s  %-4s  %s\n", "ID", "NAME", "VER", "UPDATED")
	for _, p := range resp.Msg.GetPipelines() {
		fmt.Printf("%-36s  %-30s  v%-3d  %s\n",
			p.GetPipelineId(),
			truncate(p.GetName(), 30),
			p.GetVersion(),
			p.GetUpdatedAt().AsTime().Format(time.RFC3339),
		)
	}
	return nil
}

func cmdPipelineGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl pipeline get <pipeline-id>")
	}

	client := newClient()
	resp, err := client.GetPipeline(context.Background(), withAuth(&portwhinev1.GetPipelineRequest{
		PipelineId: args[0],
	}))
	if err != nil {
		return err
	}

	out, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(resp.Msg)
	fmt.Println(string(out))
	return nil
}

func cmdPipelineCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl pipeline create <json-file>")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	def := &portwhinev1.PipelineDefinition{}
	if err := protojson.Unmarshal(data, def); err != nil {
		return fmt.Errorf("parse pipeline definition: %w", err)
	}

	client := newClient()
	resp, err := client.CreatePipeline(context.Background(), withAuth(&portwhinev1.CreatePipelineRequest{
		Definition: def,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Pipeline created: %s\n", resp.Msg.GetPipelineId())
	return nil
}

func cmdPipelineDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl pipeline delete <pipeline-id>")
	}

	client := newClient()
	_, err := client.DeletePipeline(context.Background(), withAuth(&portwhinev1.DeletePipelineRequest{
		PipelineId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Pipeline deleted.")
	return nil
}

func cmdPipelineStart(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl pipeline start <pipeline-id>")
	}

	client := newClient()
	resp, err := client.StartPipeline(context.Background(), withAuth(&portwhinev1.StartPipelineRequest{
		PipelineId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Pipeline run started: %s\n", resp.Msg.GetRunId())
	return nil
}

func cmdPipelineStop(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl pipeline stop <run-id>")
	}

	client := newClient()
	_, err := client.StopPipelineRun(context.Background(), withAuth(&portwhinev1.StopPipelineRunRequest{
		RunId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Pipeline run stopped.")
	return nil
}

// --- Run commands ---

func cmdRunList(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl run list <pipeline-id>")
	}

	client := newClient()
	resp, err := client.ListPipelineRuns(context.Background(), withAuth(&portwhinev1.ListPipelineRunsRequest{
		PipelineId: args[0],
		PageSize:   100,
	}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetRuns()) == 0 {
		fmt.Println("No runs found.")
		return nil
	}

	fmt.Printf("%-36s  %-12s  %s\n", "RUN ID", "STATE", "STARTED")
	for _, r := range resp.Msg.GetRuns() {
		started := "—"
		if r.GetStartedAt() != nil {
			started = r.GetStartedAt().AsTime().Format(time.RFC3339)
		}
		fmt.Printf("%-36s  %-12s  %s\n",
			r.GetRunId(),
			r.GetState().String(),
			started,
		)
	}
	return nil
}

func cmdRunStatus(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl run status <run-id>")
	}

	client := newClient()
	resp, err := client.GetPipelineRunStatus(context.Background(), withAuth(&portwhinev1.GetPipelineRunStatusRequest{
		RunId: args[0],
	}))
	if err != nil {
		return err
	}

	s := resp.Msg.GetStatus()

	// Header
	started := "—"
	if s.GetStartedAt() != nil {
		started = s.GetStartedAt().AsTime().Format(time.RFC3339)
	}
	finished := ""
	if s.GetFinishedAt() != nil {
		finished = "  Finished: " + s.GetFinishedAt().AsTime().Format(time.RFC3339)
	}
	fmt.Printf("Run: %s  State: %s  Started: %s%s\n\n",
		s.GetRunId(), s.GetState().String(), started, finished)

	// Node table
	nodes := s.GetNodes()
	if len(nodes) > 0 {
		fmt.Printf("%-16s %-22s %-14s %-12s %6s %6s %6s\n",
			"NODE", "TYPE", "STATUS", "CONTAINER", "IN", "OUT", "ERRORS")
		for _, n := range nodes {
			workerType := n.GetWorkerType()
			if workerType == "" {
				workerType = "—"
			}
			cStatus := n.GetContainerStatus()
			if cStatus == "" {
				cStatus = "—"
			}
			fmt.Printf("%-16s %-22s %-14s %-12s %6d %6d %6d",
				truncate(n.GetNodeId(), 16),
				truncate(workerType, 22),
				n.GetWorkerStatus().String(),
				truncate(cStatus, 12),
				n.GetItemsIn(),
				n.GetItemsOut(),
				n.GetErrors(),
			)
			if n.GetErrorMessage() != "" {
				fmt.Printf("  err: %s", truncate(n.GetErrorMessage(), 60))
			}
			if n.GetContainerMessage() != "" {
				fmt.Printf("  [%s]", truncate(n.GetContainerMessage(), 40))
			}
			fmt.Println()
		}
	} else {
		fmt.Println("No node status available.")
	}

	return nil
}

func cmdRunLogs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl run logs <run-id> <node-id> [--tail N] [--follow]")
	}

	runID := args[0]
	nodeID := args[1]
	var tail int32
	follow := false

	// Parse optional flags.
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--follow", "-f":
			follow = true
		case "--tail", "-n":
			if i+1 < len(args) {
				i++
				n, err := strconv.Atoi(args[i])
				if err == nil {
					tail = int32(n)
				}
			}
		}
	}

	client := newClient()
	stream, err := client.GetNodeLogs(context.Background(), withAuth(&portwhinev1.GetNodeLogsRequest{
		RunId:  runID,
		NodeId: nodeID,
		Tail:   tail,
		Follow: follow,
	}))
	if err != nil {
		return err
	}

	for stream.Receive() {
		fmt.Println(stream.Msg().GetLine())
	}
	if err := stream.Err(); err != nil {
		return err
	}
	return nil
}

func cmdRunStream(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl run stream <run-id>")
	}

	client := newClient()
	stream, err := client.StreamPipelineResults(context.Background(), withAuth(&portwhinev1.StreamPipelineResultsRequest{
		RunId: args[0],
	}))
	if err != nil {
		return err
	}

	for stream.Receive() {
		item := stream.Msg().GetItem()
		if item == nil {
			continue
		}
		out, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(item)
		fmt.Println(string(out))
	}
	if err := stream.Err(); err != nil {
		return err
	}
	return nil
}

// --- Worker commands ---

func cmdWorkerList() error {
	client := newClient()
	resp, err := client.ListWorkerImages(context.Background(), withAuth(&portwhinev1.ListWorkerImagesRequest{}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetImages()) == 0 {
		fmt.Println("No worker images registered.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %s\n", "ID", "NAME", "IMAGE")
	for _, img := range resp.Msg.GetImages() {
		fmt.Printf("%-36s  %-20s  %s\n", img.GetId(), img.GetName(), img.GetImage())
	}
	return nil
}

func cmdWorkerRegister(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl worker register <name> <image> [description]")
	}

	desc := ""
	if len(args) >= 3 {
		desc = args[2]
	}

	client := newClient()
	resp, err := client.RegisterWorkerImage(context.Background(), withAuth(&portwhinev1.RegisterWorkerImageRequest{
		Name:        args[0],
		Image:       args[1],
		Description: desc,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Worker image registered: %s\n", resp.Msg.GetWorkerImageId())
	return nil
}

func cmdWorkerDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl worker delete <worker-image-id>")
	}

	client := newClient()
	_, err := client.DeleteWorkerImage(context.Background(), withAuth(&portwhinev1.DeleteWorkerImageRequest{
		WorkerImageId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Worker image deleted.")
	return nil
}

// --- User commands ---

func cmdUserList() error {
	client := newClient()
	resp, err := client.ListUsers(context.Background(), withAuth(&portwhinev1.ListUsersRequest{
		PageSize: 100,
	}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetUsers()) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-30s  %-8s  %s\n", "ID", "USERNAME", "EMAIL", "ROLE", "ACTIVE")
	for _, u := range resp.Msg.GetUsers() {
		fmt.Printf("%-36s  %-20s  %-30s  %-8s  %v\n",
			u.GetId(), u.GetUsername(), u.GetEmail(), u.GetRole(), u.GetIsActive())
	}
	return nil
}

func cmdUserGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl user get <user-id>")
	}

	client := newClient()
	resp, err := client.GetUser(context.Background(), withAuth(&portwhinev1.GetUserRequest{
		UserId: args[0],
	}))
	if err != nil {
		return err
	}

	out, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(resp.Msg)
	fmt.Println(string(out))
	return nil
}

func cmdUserDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl user delete <user-id>")
	}

	client := newClient()
	_, err := client.DeleteUser(context.Background(), withAuth(&portwhinev1.DeleteUserRequest{
		UserId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("User deleted.")
	return nil
}

// --- API Key commands ---

func cmdAPIKeyList() error {
	client := newClient()
	resp, err := client.ListAPIKeys(context.Background(), withAuth(&portwhinev1.ListAPIKeysRequest{}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetKeys()) == 0 {
		fmt.Println("No API keys found.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-10s  %-8s  %s\n", "ID", "NAME", "PREFIX", "REVOKED", "CREATED")
	for _, k := range resp.Msg.GetKeys() {
		fmt.Printf("%-36s  %-20s  %-10s  %-8v  %s\n",
			k.GetId(), k.GetName(), k.GetKeyPrefix(), k.GetRevoked(),
			k.GetCreatedAt().AsTime().Format(time.RFC3339))
	}
	return nil
}

func cmdAPIKeyCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl apikey create <name>")
	}

	client := newClient()
	resp, err := client.CreateAPIKey(context.Background(), withAuth(&portwhinev1.CreateAPIKeyRequest{
		Name: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Printf("API key created (save this - it will not be shown again):\n")
	fmt.Printf("  Key:    %s\n", resp.Msg.GetApiKey())
	fmt.Printf("  Prefix: %s\n", resp.Msg.GetKeyPrefix())
	return nil
}

func cmdAPIKeyRevoke(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl apikey revoke <api-key-id>")
	}

	client := newClient()
	_, err := client.RevokeAPIKey(context.Background(), withAuth(&portwhinev1.RevokeAPIKeyRequest{
		ApiKeyId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("API key revoked.")
	return nil
}

// --- Team commands ---

func dispatchTeam(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl team <list|get|create|delete|members|add-member|remove-member|update-role>")
	}
	switch args[0] {
	case "list":
		return cmdTeamList()
	case "get":
		return cmdTeamGet(args[1:])
	case "create":
		return cmdTeamCreate(args[1:])
	case "delete":
		return cmdTeamDelete(args[1:])
	case "members":
		return cmdTeamMembers(args[1:])
	case "add-member":
		return cmdTeamAddMember(args[1:])
	case "remove-member":
		return cmdTeamRemoveMember(args[1:])
	case "update-role":
		return cmdTeamUpdateRole(args[1:])
	default:
		return fmt.Errorf("unknown team subcommand: %s", args[0])
	}
}

func cmdTeamList() error {
	client := newClient()
	resp, err := client.ListTeams(context.Background(), withAuth(&portwhinev1.ListTeamsRequest{
		PageSize: 100,
	}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetTeams()) == 0 {
		fmt.Println("No teams found.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-8s  %s\n", "ID", "NAME", "MEMBERS", "CREATED")
	for _, t := range resp.Msg.GetTeams() {
		fmt.Printf("%-36s  %-20s  %-8d  %s\n",
			t.GetId(), truncate(t.GetName(), 20), t.GetMemberCount(),
			t.GetCreatedAt().AsTime().Format(time.RFC3339))
	}
	return nil
}

func cmdTeamGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl team get <team-id>")
	}

	client := newClient()
	resp, err := client.GetTeam(context.Background(), withAuth(&portwhinev1.GetTeamRequest{
		TeamId: args[0],
	}))
	if err != nil {
		return err
	}

	out, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(resp.Msg)
	fmt.Println(string(out))
	return nil
}

func cmdTeamCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl team create <name> [description]")
	}

	desc := ""
	if len(args) >= 2 {
		desc = strings.Join(args[1:], " ")
	}

	client := newClient()
	resp, err := client.CreateTeam(context.Background(), withAuth(&portwhinev1.CreateTeamRequest{
		Name:        args[0],
		Description: desc,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Team created: %s\n", resp.Msg.GetTeamId())
	return nil
}

func cmdTeamDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl team delete <team-id>")
	}

	client := newClient()
	_, err := client.DeleteTeam(context.Background(), withAuth(&portwhinev1.DeleteTeamRequest{
		TeamId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Team deleted.")
	return nil
}

func cmdTeamMembers(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl team members <team-id>")
	}

	client := newClient()
	resp, err := client.ListTeamMembers(context.Background(), withAuth(&portwhinev1.ListTeamMembersRequest{
		TeamId: args[0],
	}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetMembers()) == 0 {
		fmt.Println("No members found.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-30s  %-10s  %s\n", "USER ID", "USERNAME", "EMAIL", "ROLE", "JOINED")
	for _, m := range resp.Msg.GetMembers() {
		fmt.Printf("%-36s  %-20s  %-30s  %-10s  %s\n",
			m.GetUserId(), m.GetUsername(), m.GetEmail(), m.GetRole(),
			m.GetJoinedAt().AsTime().Format(time.RFC3339))
	}
	return nil
}

func cmdTeamAddMember(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl team add-member <team-id> <user-id> [role]")
	}

	role := "member"
	if len(args) >= 3 {
		role = args[2]
	}

	client := newClient()
	_, err := client.AddTeamMember(context.Background(), withAuth(&portwhinev1.AddTeamMemberRequest{
		TeamId: args[0],
		UserId: args[1],
		Role:   role,
	}))
	if err != nil {
		return err
	}

	fmt.Println("Team member added.")
	return nil
}

func cmdTeamRemoveMember(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl team remove-member <team-id> <user-id>")
	}

	client := newClient()
	_, err := client.RemoveTeamMember(context.Background(), withAuth(&portwhinev1.RemoveTeamMemberRequest{
		TeamId: args[0],
		UserId: args[1],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Team member removed.")
	return nil
}

func cmdTeamUpdateRole(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: pwctl team update-role <team-id> <user-id> <role>")
	}

	client := newClient()
	_, err := client.UpdateTeamMemberRole(context.Background(), withAuth(&portwhinev1.UpdateTeamMemberRoleRequest{
		TeamId: args[0],
		UserId: args[1],
		Role:   args[2],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Team member role updated.")
	return nil
}

// --- Permission commands ---

func dispatchPermission(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl permission <grant|revoke|list|my>")
	}
	switch args[0] {
	case "grant":
		return cmdPermissionGrant(args[1:])
	case "revoke":
		return cmdPermissionRevoke(args[1:])
	case "list":
		return cmdPermissionList(args[1:])
	case "my":
		return cmdPermissionMy()
	default:
		return fmt.Errorf("unknown permission subcommand: %s", args[0])
	}
}

func cmdPermissionGrant(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: pwctl permission grant <subject-type> <subject-id> <resource-type> <resource-id> <action> [effect]\n  subject-type: user, team\n  action: read, update, delete, execute, *\n  effect: allow (default), deny")
	}

	effect := "allow"
	if len(args) >= 6 {
		effect = args[5]
	}

	client := newClient()
	resp, err := client.GrantPermission(context.Background(), withAuth(&portwhinev1.GrantPermissionRequest{
		SubjectType:  args[0],
		SubjectId:    args[1],
		ResourceType: args[2],
		ResourceId:   args[3],
		Action:       args[4],
		Effect:       effect,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Permission granted: %s\n", resp.Msg.GetPermissionId())
	return nil
}

func cmdPermissionRevoke(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl permission revoke <permission-id>")
	}

	client := newClient()
	_, err := client.RevokePermission(context.Background(), withAuth(&portwhinev1.RevokePermissionRequest{
		PermissionId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Permission revoked.")
	return nil
}

func cmdPermissionList(args []string) error {
	req := &portwhinev1.ListPermissionsRequest{}

	// Parse optional filters: --subject-type, --subject-id, --resource-type, --resource-id
	for i := 0; i < len(args)-1; i += 2 {
		switch args[i] {
		case "--subject-type":
			req.SubjectType = args[i+1]
		case "--subject-id":
			req.SubjectId = args[i+1]
		case "--resource-type":
			req.ResourceType = args[i+1]
		case "--resource-id":
			req.ResourceId = args[i+1]
		}
	}

	client := newClient()
	resp, err := client.ListPermissions(context.Background(), withAuth(req))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetPermissions()) == 0 {
		fmt.Println("No permissions found.")
		return nil
	}

	fmt.Printf("%-36s  %-6s  %-36s  %-12s  %-36s  %-8s  %s\n",
		"ID", "TYPE", "SUBJECT", "RESOURCE", "RESOURCE ID", "ACTION", "EFFECT")
	for _, p := range resp.Msg.GetPermissions() {
		fmt.Printf("%-36s  %-6s  %-36s  %-12s  %-36s  %-8s  %s\n",
			p.GetId(), p.GetSubjectType(), truncate(p.GetSubjectId(), 36),
			p.GetResourceType(), truncate(p.GetResourceId(), 36),
			p.GetAction(), p.GetEffect())
	}
	return nil
}

func cmdPermissionMy() error {
	client := newClient()
	resp, err := client.ListMyPermissions(context.Background(), withAuth(&portwhinev1.ListMyPermissionsRequest{}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetPermissions()) == 0 {
		fmt.Println("No permissions found.")
		return nil
	}

	fmt.Printf("%-36s  %-6s  %-12s  %-36s  %-8s  %s\n",
		"ID", "TYPE", "RESOURCE", "RESOURCE ID", "ACTION", "EFFECT")
	for _, p := range resp.Msg.GetPermissions() {
		fmt.Printf("%-36s  %-6s  %-12s  %-36s  %-8s  %s\n",
			p.GetId(), p.GetSubjectType(), p.GetResourceType(),
			truncate(p.GetResourceId(), 36), p.GetAction(), p.GetEffect())
	}
	return nil
}

// --- Role commands ---

func dispatchRole(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: pwctl role <list|create|update|delete>")
	}
	switch args[0] {
	case "list":
		return cmdRoleList()
	case "create":
		return cmdRoleCreate(args[1:])
	case "update":
		return cmdRoleUpdate(args[1:])
	case "delete":
		return cmdRoleDelete(args[1:])
	default:
		return fmt.Errorf("unknown role subcommand: %s", args[0])
	}
}

func cmdRoleList() error {
	client := newClient()
	resp, err := client.ListRoles(context.Background(), withAuth(&portwhinev1.ListRolesRequest{}))
	if err != nil {
		return err
	}

	if len(resp.Msg.GetRoles()) == 0 {
		fmt.Println("No roles found.")
		return nil
	}

	fmt.Printf("%-36s  %-15s  %-8s  %-8s  %s\n", "ID", "NAME", "SYSTEM", "CUSTOM", "DESCRIPTION")
	for _, r := range resp.Msg.GetRoles() {
		fmt.Printf("%-36s  %-15s  %-8v  %-8v  %s\n",
			r.GetId(), r.GetName(), r.GetIsSystem(), r.GetIsCustom(),
			truncate(r.GetDescription(), 40))
	}
	return nil
}

func cmdRoleCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl role create <name> [description]")
	}

	desc := ""
	if len(args) >= 2 {
		desc = strings.Join(args[1:], " ")
	}

	client := newClient()
	resp, err := client.CreateRole(context.Background(), withAuth(&portwhinev1.CreateRoleRequest{
		Name:        args[0],
		Description: desc,
	}))
	if err != nil {
		return err
	}

	fmt.Printf("Role created: %s\n", resp.Msg.GetRoleId())
	return nil
}

func cmdRoleUpdate(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pwctl role update <role-id> <name> [description]")
	}

	desc := ""
	if len(args) >= 3 {
		desc = strings.Join(args[2:], " ")
	}

	client := newClient()
	_, err := client.UpdateRole(context.Background(), withAuth(&portwhinev1.UpdateRoleRequest{
		RoleId:      args[0],
		Name:        args[1],
		Description: desc,
	}))
	if err != nil {
		return err
	}

	fmt.Println("Role updated.")
	return nil
}

func cmdRoleDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: pwctl role delete <role-id>")
	}

	client := newClient()
	_, err := client.DeleteRole(context.Background(), withAuth(&portwhinev1.DeleteRoleRequest{
		RoleId: args[0],
	}))
	if err != nil {
		return err
	}

	fmt.Println("Role deleted.")
	return nil
}

// --- Helpers ---

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

