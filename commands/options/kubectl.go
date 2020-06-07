package options

import "github.com/spf13/cobra"

// Kubectl struct contains options regarding the Kubernetes context and namespace
type Kubectl struct {
	Context   string
	Namespace string
}

func AddContextArg(cmd *cobra.Command, k *Kubectl) {
	cmd.PersistentFlags().StringVarP(&k.Context, "context", "c", "", "Kubernetes context")
}

func AddNamespaceArg(cmd *cobra.Command, k *Kubectl) {
	cmd.PersistentFlags().StringVarP(&k.Namespace, "namespace", "n", "default", "Kubernetes namespace")
	must(cmd.MarkPersistentFlagRequired("namespace"))
}
