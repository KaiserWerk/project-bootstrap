package gittools

import (
	"context"
	"log"
	"os/exec"
)

func Clone(ctx context.Context, repoURL, workDir, destDir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, destDir)
	cmd.Dir = workDir
	log.Printf("Executing command: %s\n", cmd.String())
	return cmd.Run()
}

func Add(ctx context.Context, repoDir string) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = repoDir
	log.Printf("Executing command: %s\n", cmd.String())
	return cmd.Run()
}

func Commit(ctx context.Context, repoDir, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", `"`+message+`"`)
	cmd.Dir = repoDir
	log.Printf("Executing command: %s\n", cmd.String())
	return cmd.Run()
}

func Push(ctx context.Context, repoDir string) error {
	cmd := exec.CommandContext(ctx, "git", "push")
	cmd.Dir = repoDir
	log.Printf("Executing command: %s\n", cmd.String())
	return cmd.Run()
}

func CheckoutNew(ctx context.Context, repoDir, branchName string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = repoDir
	log.Printf("Executing command: %s\n", cmd.String())
	return cmd.Run()
}
