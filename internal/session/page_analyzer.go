package session

import (
	"encoding/json"
	"fmt"
)

// PageStructure represents the analyzed structure of a web page
type PageStructure struct {
	PageID    string          `json:"page_id"`
	URL       string          `json:"url"`
	Title     string          `json:"title"`
	Structure StructureDetail `json:"structure"`
}

// StructureDetail contains the extracted page structure elements
type StructureDetail struct {
	Classes          []string            `json:"classes"`
	IDs              []string            `json:"ids"`
	Headings         map[string][]string `json:"headings"`
	Interactive      InteractiveDetail   `json:"interactive"`
	SemanticSections []SemanticSection   `json:"semantic_sections"`
	DataAttributes   []string            `json:"data_attributes"`
	TextSnippets     []string            `json:"text_snippets"`
}

// InteractiveDetail contains interactive element summaries
type InteractiveDetail struct {
	Buttons []string `json:"buttons"`
	Links   []string `json:"links"`
	Forms   []string `json:"forms"`
}

// SemanticSection represents a semantic HTML section
type SemanticSection struct {
	Type     string   `json:"type"`
	Class    string   `json:"class,omitempty"`
	Count    int      `json:"count"`
	Children []string `json:"children,omitempty"`
}

// pageAnalyzerJS is the JavaScript code that extracts page structure.
// It returns a JSON-serializable object matching the PageStructure Go type.
const pageAnalyzerJS = `(function() {
  var result = {
    url: location.href,
    title: document.title,
    structure: {
      classes: [],
      ids: [],
      headings: {},
      interactive: { buttons: [], links: [], forms: [] },
      semantic_sections: [],
      data_attributes: [],
      text_snippets: []
    }
  };

  // Extract unique CSS classes
  var classSet = {};
  document.querySelectorAll('[class]').forEach(function(el) {
    el.classList.forEach(function(c) { classSet[c] = true; });
  });
  result.structure.classes = Object.keys(classSet).sort().map(function(c) { return '.' + c; });

  // Extract unique IDs
  var ids = [];
  document.querySelectorAll('[id]').forEach(function(el) {
    ids.push('#' + el.id);
  });
  result.structure.ids = ids;

  // Extract headings h1-h6
  ['h1','h2','h3','h4','h5','h6'].forEach(function(tag) {
    var els = document.querySelectorAll(tag);
    if (els.length > 0) {
      result.structure.headings[tag] = Array.from(els).map(function(el) {
        return el.textContent.trim().substring(0, 100);
      });
    }
  });

  // Extract interactive elements - buttons
  var btnMap = {};
  document.querySelectorAll('button, [role="button"], input[type="button"], input[type="submit"]').forEach(function(el) {
    var key = el.className ? '.' + el.className.split(/\s+/)[0] : el.tagName.toLowerCase();
    btnMap[key] = (btnMap[key] || 0) + 1;
  });
  result.structure.interactive.buttons = Object.keys(btnMap).map(function(k) {
    return k + ' (' + btnMap[k] + ')';
  });

  // Extract interactive elements - links
  var linkMap = {};
  document.querySelectorAll('a[href]').forEach(function(el) {
    var key = el.className ? '.' + el.className.split(/\s+/)[0] : 'a';
    linkMap[key] = (linkMap[key] || 0) + 1;
  });
  result.structure.interactive.links = Object.keys(linkMap).map(function(k) {
    return k + ' (' + linkMap[k] + ')';
  });

  // Extract interactive elements - forms
  var formMap = {};
  document.querySelectorAll('form').forEach(function(el) {
    var key = el.className ? '.' + el.className.split(/\s+/)[0] : 'form';
    var inputs = el.querySelectorAll('input, select, textarea').length;
    formMap[key] = { count: (formMap[key] ? formMap[key].count : 0) + 1, inputs: inputs };
  });
  result.structure.interactive.forms = Object.keys(formMap).map(function(k) {
    return k + ' (' + formMap[k].count + ', ' + formMap[k].inputs + ' inputs)';
  });

  // Extract semantic sections
  ['article','nav','section','main','aside','header','footer'].forEach(function(tag) {
    var els = document.querySelectorAll(tag);
    if (els.length === 0) return;

    // Group by class
    var groups = {};
    els.forEach(function(el) {
      var cls = el.className ? el.className.split(/\s+/)[0] : '';
      var key = cls || '_noclass';
      if (!groups[key]) {
        groups[key] = { count: 0, childTags: {} };
      }
      groups[key].count++;
      // Sample children from first element of this group
      if (groups[key].count === 1) {
        Array.from(el.children).forEach(function(child) {
          var childKey = child.tagName.toLowerCase();
          if (child.className) childKey += '.' + child.className.split(/\s+/)[0];
          groups[key].childTags[childKey] = true;
        });
      }
    });

    Object.keys(groups).forEach(function(cls) {
      var g = groups[cls];
      var section = {
        type: tag,
        count: g.count,
        children: Object.keys(g.childTags).slice(0, 10)
      };
      if (cls !== '_noclass') section['class'] = cls;
      result.structure.semantic_sections.push(section);
    });
  });

  // Extract data attributes
  var dataAttrSet = {};
  document.querySelectorAll('*').forEach(function(el) {
    Array.from(el.attributes).forEach(function(attr) {
      if (attr.name.indexOf('data-') === 0) {
        dataAttrSet[attr.name] = true;
      }
    });
  });
  result.structure.data_attributes = Object.keys(dataAttrSet).sort();

  // Extract text snippets from major blocks
  var snippets = [];
  document.querySelectorAll('p, li, td, h1, h2, h3, blockquote').forEach(function(el) {
    var text = el.textContent.trim();
    if (text.length > 10 && snippets.length < 20) {
      snippets.push(text.substring(0, 50));
    }
  });
  result.structure.text_snippets = snippets;

  return result;
})();`

// AnalyzePage extracts the structural overview of a page.
// Results are cached per pageID — call InvalidatePageAnalysis to clear.
func (s *Session) AnalyzePage(targetID string) (*PageStructure, error) {
	// Check cache first
	if s.pageAnalysisCache != nil {
		if cached, ok := s.pageAnalysisCache[targetID]; ok {
			return cached, nil
		}
	}

	// Execute the analyzer JavaScript
	result, err := s.ExecuteJavascript(targetID, pageAnalyzerJS)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze page: %w", err)
	}

	// The result comes back as a map[string]interface{} from Runtime.evaluate
	// Marshal back to JSON then unmarshal into our typed struct
	rawJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis result: %w", err)
	}

	var structure PageStructure
	if err := json.Unmarshal(rawJSON, &structure); err != nil {
		return nil, fmt.Errorf("failed to parse analysis result: %w", err)
	}

	// Set the page ID (not available from JavaScript)
	structure.PageID = targetID

	// Cache the result
	if s.pageAnalysisCache == nil {
		s.pageAnalysisCache = make(map[string]*PageStructure)
	}
	s.pageAnalysisCache[targetID] = &structure

	return &structure, nil
}

// InvalidatePageAnalysis clears the cached analysis for a specific page
func (s *Session) InvalidatePageAnalysis(pageID string) {
	if s.pageAnalysisCache != nil {
		delete(s.pageAnalysisCache, pageID)
	}
}

// InvalidateAllPageAnalysis clears all cached page analyses
func (s *Session) InvalidateAllPageAnalysis() {
	s.pageAnalysisCache = make(map[string]*PageStructure)
}
