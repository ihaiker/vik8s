package pod

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	AffinityExecution struct {
		name        string
		fields      [][]string
		expressions [][]string
		weight      string
		namespaces  []string
	}
)

func executionParse(weights *config.Directive) []AffinityExecution {
	executions := make([]AffinityExecution, 0)

	selector(weights, func(weight *config.Directive) {
		ae := AffinityExecution{
			fields: make([][]string, 0), expressions: make([][]string, 0),
		}
		ae.name, ae.weight = utils.Split2(weight.Name, ":")

		selector(weight, func(execution *config.Directive) {
			switch execution.Name {
			case "matchFields", "fields", "labels":
				if len(execution.Args) == 0 {
					for _, fieldsBody := range execution.Body {
						ae.fields = append(ae.fields, append([]string{fieldsBody.Name}, fieldsBody.Args...))
					}
				} else {
					ae.fields = append(ae.fields, execution.Args)
				}
			case "matchExpressions", "expr":
				if len(execution.Args) == 0 {
					for _, fieldsBody := range execution.Body {
						ae.expressions = append(ae.expressions, append([]string{fieldsBody.Name}, fieldsBody.Args...))
					}
				} else {
					ae.expressions = append(ae.expressions, execution.Args)
				}
			case "namespaces":
				ae.namespaces = execution.Args
			}
		})
		executions = append(executions, ae)
	})
	return executions
}

func AffinityNodeParse(nodeAffinity *v1.NodeAffinity, nodeConfig *config.Directive) {
	selector(nodeConfig, func(body *config.Directive) {
		switch body.Name {
		case "preferred":
			exprs := executionParse(body)
			for _, expr := range exprs {

				psst := v1.PreferredSchedulingTerm{
					Weight: *utils.Int32(expr.weight, 10),
				}

				for _, field := range expr.fields {
					nsr := v1.NodeSelectorRequirement{
						Key: field[0], Operator: v1.NodeSelectorOperator(field[1]), Values: field[2:],
					}
					psst.Preference.MatchFields = append(psst.Preference.MatchFields, nsr)
				}

				for _, field := range expr.expressions {
					nsr := v1.NodeSelectorRequirement{
						Key: field[0], Operator: v1.NodeSelectorOperator(field[1]), Values: field[2:],
					}
					psst.Preference.MatchExpressions = append(psst.Preference.MatchExpressions, nsr)
				}

				nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
					append(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, psst)
			}
		case "required":
			exprs := executionParse(body)
			ns := &v1.NodeSelector{}
			for _, expr := range exprs {
				nst := v1.NodeSelectorTerm{}
				for _, field := range expr.fields {
					nsr := v1.NodeSelectorRequirement{
						Key: field[0], Operator: v1.NodeSelectorOperator(field[1]), Values: field[2:],
					}
					nst.MatchFields = append(nst.MatchFields, nsr)
				}

				for _, field := range expr.expressions {
					nsr := v1.NodeSelectorRequirement{
						Key: field[0], Operator: v1.NodeSelectorOperator(field[1]), Values: field[2:],
					}
					nst.MatchExpressions = append(nst.MatchExpressions, nsr)
				}
				ns.NodeSelectorTerms = append(ns.NodeSelectorTerms, nst)
			}
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = ns
		}
	})
}

func affinityPodParse(podConfig *config.Directive) (Preferred []v1.WeightedPodAffinityTerm, Required []v1.PodAffinityTerm) {
	selector(podConfig, func(body *config.Directive) {
		switch body.Name {
		case "preferred":
			exprs := executionParse(body)
			for _, expr := range exprs {
				psst := v1.WeightedPodAffinityTerm{
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: make(map[string]string),
						},
					},
					Weight: *utils.Int32(expr.weight, 10),
				}
				for _, field := range expr.fields {
					psst.PodAffinityTerm.LabelSelector.MatchLabels[field[0]] = field[1]
				}
				for _, field := range expr.expressions {
					nsr := metav1.LabelSelectorRequirement{
						Key: field[0], Operator: metav1.LabelSelectorOperator(field[1]), Values: field[2:],
					}
					psst.PodAffinityTerm.LabelSelector.MatchExpressions = append(psst.PodAffinityTerm.LabelSelector.MatchExpressions, nsr)
				}
				Preferred = append(Preferred, psst)
			}
		case "required":
			exprs := executionParse(body)
			for _, expr := range exprs {
				ns := v1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: make(map[string]string),
					}, Namespaces: expr.namespaces, TopologyKey: expr.name,
				}
				for _, field := range expr.fields {
					ns.LabelSelector.MatchLabels[field[0]] = field[1]
				}
				for _, field := range expr.expressions {
					nsr := metav1.LabelSelectorRequirement{
						Key: field[0], Operator: metav1.LabelSelectorOperator(field[1]), Values: field[2:],
					}
					ns.LabelSelector.MatchExpressions = append(ns.LabelSelector.MatchExpressions, nsr)
				}
				Required = append(Required, ns)
			}
		}
	})
	return
}

func AffinityParse(d *config.Directive, spec *v1.PodSpec) {
	if spec.Affinity == nil {
		spec.Affinity = &v1.Affinity{}
	}
	selector(d, func(body *config.Directive) {
		switch body.Name {
		case "node":
			if spec.Affinity.NodeAffinity == nil {
				spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
			}
			AffinityNodeParse(spec.Affinity.NodeAffinity, body)

		case "pod":
			if spec.Affinity.PodAffinity == nil {
				spec.Affinity.PodAffinity = &v1.PodAffinity{}
			}
			spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = affinityPodParse(body)

		case "podAnti":
			if spec.Affinity.PodAntiAffinity == nil {
				spec.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{}
			}
			spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = affinityPodParse(body)
		}
	})
}

func selector(d *config.Directive, comsumer func(*config.Directive)) {
	if len(d.Args) == 0 {
		for _, body := range d.Body {
			comsumer(&config.Directive{
				Name: body.Name, Args: body.Args,
				Body: body.Body,
			})
		}
	} else {
		comsumer(&config.Directive{
			Name: d.Args[0], Args: d.Args[1:], Body: d.Body,
		})
	}
}
