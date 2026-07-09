package service

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type RelationValidationIssue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type GraphContextPreview struct {
	PrimaryParentChain []string `json:"primary_parent_chain,omitempty"`
	EnvironmentChain   []string `json:"environment_chain,omitempty"`
	OrganizationChains []string `json:"organization_chains,omitempty"`
	SocialLinks        []string `json:"social_links,omitempty"`
	AuxiliaryScopes    []string `json:"auxiliary_scopes,omitempty"`
	Summary            []string `json:"summary,omitempty"`
}

func BuildNodeDiagnostics(node *store.NodeModel, relations []store.RelationModel) ([]RelationValidationIssue, *GraphContextPreview) {
	if node == nil {
		return nil, &GraphContextPreview{}
	}
	nameByID := buildNodeNameMap(node, relations)
	outgoing := make([]store.RelationModel, 0, len(relations))
	for _, rel := range relations {
		if rel.SourceUUID == node.UUID {
			outgoing = append(outgoing, rel)
		}
	}

	issues := buildRelationValidationIssues(node, outgoing)
	preview := buildGraphContextPreview(node, outgoing, nameByID)
	return issues, preview
}

func buildRelationValidationIssues(node *store.NodeModel, outgoing []store.RelationModel) []RelationValidationIssue {
	issues := make([]RelationValidationIssue, 0)
	locatedAtCount := 0
	externalParentCount := 0
	parentUUID := ""
	if node.ParentUUID != nil {
		parentUUID = *node.ParentUUID
	}

	for _, rel := range outgoing {
		switch engine.RelationType(rel.RelationType) {
		case engine.RelLocatedAt:
			locatedAtCount++
			if rel.TargetUUID == parentUUID && parentUUID != "" {
				issues = append(issues, RelationValidationIssue{
					Severity: "warning",
					Code:     "located_at_matches_primary_parent",
					Message:  "located_at points to the same node as Primary Parent. Keep parent for stable hierarchy and reserve located_at for temporary environment position.",
				})
			}
		case engine.RelBelongsTo, engine.RelSubordinate:
			if rel.TargetUUID == parentUUID && parentUUID != "" {
				issues = append(issues, RelationValidationIssue{
					Severity: "info",
					Code:     "organization_relation_matches_primary_parent",
					Message:  rel.RelationType + " points to the same node as Primary Parent. This is allowed, but parent changes stable hierarchy while organization edges only add affiliation/control semantics.",
				})
			}
		case engine.RelExternalParent:
			externalParentCount++
			issues = append(issues, RelationValidationIssue{
				Severity: "info",
				Code:     "external_parent_auxiliary_scope",
				Message:  "external_parent is auxiliary DAG scope only. It does not enter default context assembly or default propagation.",
			})
		case engine.RelAlly, engine.RelEnemy, engine.RelKinship:
			issues = append(issues, RelationValidationIssue{
				Severity: "info",
				Code:     "social_relation_background_only",
				Message:  rel.RelationType + " is a background social edge and is not part of the default hierarchy/environment traversal path.",
			})
		}
	}

	if locatedAtCount > 1 {
		issues = append(issues, RelationValidationIssue{
			Severity: "warning",
			Code:     "multiple_located_at_edges",
			Message:  "More than one located_at edge is active on this node. Prefer a single current environment unless simultaneous positions are explicitly intended.",
		})
	}
	if externalParentCount > 1 {
		issues = append(issues, RelationValidationIssue{
			Severity: "info",
			Code:     "multiple_external_parent_edges",
			Message:  "Multiple external_parent edges are present. Keep auxiliary DAG scope sparse so it does not become a hidden replacement for the primary hierarchy.",
		})
	}
	if node.NodeType == string(engine.NodeTypeNPC) && locatedAtCount == 0 {
		issues = append(issues, RelationValidationIssue{
			Severity: "info",
			Code:     "npc_missing_located_at",
			Message:  "NPC has no located_at edge. If this character can move between scenes, model current position with located_at instead of rewriting the parent tree.",
		})
	}

	return issues
}

func buildGraphContextPreview(node *store.NodeModel, outgoing []store.RelationModel, nameByID map[string]string) *GraphContextPreview {
	preview := &GraphContextPreview{}
	preview.PrimaryParentChain = buildParentChain(node, nameByID)
	preview.EnvironmentChain = buildScopedChains(node, outgoing, nameByID, map[string]bool{string(engine.RelLocatedAt): true})
	preview.OrganizationChains = buildScopedChains(node, outgoing, nameByID, map[string]bool{string(engine.RelBelongsTo): true, string(engine.RelSubordinate): true})
	preview.AuxiliaryScopes = buildScopedChains(node, outgoing, nameByID, map[string]bool{string(engine.RelExternalParent): true})
	preview.SocialLinks = buildSocialLinks(outgoing, nameByID)

	if len(preview.PrimaryParentChain) > 0 {
		preview.Summary = append(preview.Summary, "identity: "+strings.Join(preview.PrimaryParentChain, " > "))
	}
	if len(preview.EnvironmentChain) > 0 {
		preview.Summary = append(preview.Summary, "environment: "+strings.Join(preview.EnvironmentChain, " | "))
	}
	if len(preview.OrganizationChains) > 0 {
		preview.Summary = append(preview.Summary, "organization: "+strings.Join(preview.OrganizationChains, " | "))
	}
	if len(preview.AuxiliaryScopes) > 0 {
		preview.Summary = append(preview.Summary, "auxiliary: "+strings.Join(preview.AuxiliaryScopes, " | "))
	}
	if len(preview.SocialLinks) > 0 {
		preview.Summary = append(preview.Summary, "social: "+strings.Join(preview.SocialLinks, " | "))
	}
	return preview
}

func buildNodeNameMap(node *store.NodeModel, relations []store.RelationModel) map[string]string {
	ids := []string{node.UUID}
	if node.ParentUUID != nil {
		ids = append(ids, *node.ParentUUID)
	}
	for _, rel := range relations {
		if rel.SourceUUID != "" {
			ids = append(ids, rel.SourceUUID)
		}
		if rel.TargetUUID != "" {
			ids = append(ids, rel.TargetUUID)
		}
	}
	nodes, err := store.FindNodesByIDs(uniqueStrings(ids))
	if err != nil {
		return map[string]string{node.UUID: node.Name}
	}
	result := map[string]string{}
	for _, item := range nodes {
		result[item.UUID] = fmt.Sprintf("%s(%s)", item.Name, item.NodeType)
	}
	if _, ok := result[node.UUID]; !ok {
		result[node.UUID] = fmt.Sprintf("%s(%s)", node.Name, node.NodeType)
	}
	return result
}

func buildParentChain(node *store.NodeModel, nameByID map[string]string) []string {
	chain := []string{displayNodeRef(node.UUID, nameByID)}
	visited := map[string]bool{node.UUID: true}
	parentUUID := ""
	if node.ParentUUID != nil {
		parentUUID = *node.ParentUUID
	}
	for parentUUID != "" && !visited[parentUUID] {
		visited[parentUUID] = true
		chain = append(chain, displayNodeRef(parentUUID, nameByID))
		parentNode, err := store.GetNode(parentUUID)
		if err != nil || parentNode.ParentUUID == nil {
			break
		}
		parentUUID = *parentNode.ParentUUID
	}
	return chain
}

func buildScopedChains(node *store.NodeModel, outgoing []store.RelationModel, nameByID map[string]string, allowed map[string]bool) []string {
	var chains []string
	for _, rel := range outgoing {
		if !allowed[rel.RelationType] {
			continue
		}
		segments := []string{rel.RelationType + ": " + displayNodeRef(rel.TargetUUID, nameByID)}
		ancestorNode, err := store.GetNode(rel.TargetUUID)
		if err == nil && ancestorNode.ParentUUID != nil {
			visited := map[string]bool{rel.TargetUUID: true}
			parentUUID := *ancestorNode.ParentUUID
			for parentUUID != "" && !visited[parentUUID] {
				visited[parentUUID] = true
				segments = append(segments, displayNodeRef(parentUUID, nameByID))
				parentNode, err := store.GetNode(parentUUID)
				if err != nil || parentNode.ParentUUID == nil {
					break
				}
				parentUUID = *parentNode.ParentUUID
			}
		}
		chains = append(chains, strings.Join(segments, " > "))
	}
	return chains
}

func buildSocialLinks(outgoing []store.RelationModel, nameByID map[string]string) []string {
	var links []string
	for _, rel := range outgoing {
		switch engine.RelationType(rel.RelationType) {
		case engine.RelAlly, engine.RelEnemy, engine.RelKinship:
			links = append(links, rel.RelationType+": "+displayNodeRef(rel.TargetUUID, nameByID))
		}
	}
	return links
}

func displayNodeRef(nodeUUID string, nameByID map[string]string) string {
	if value, ok := nameByID[nodeUUID]; ok && value != "" {
		return value
	}
	return nodeUUID
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
