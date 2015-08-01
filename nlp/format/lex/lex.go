package lex

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"yap/alg/graph"
	"yap/nlp/types"
)

const (
	APPROX_LEX_SIZE         = 100000
	SEPARATOR               = " "
	MSR_SEPARATOR           = ":"
	FEATURE_SEPARATOR       = "-"
	PREFIX_SEPARATOR        = "^"
	PREFIX_MSR_SEPARATOR    = "+"
	FEATURE_PAIR_SEPARATOR  = "|"
	FEATURE_VALUE_SEPARATOR = "="
)

var (
	SKIP_BINYAN         = true
	MSR_TYPE_FROM_VALUE = map[string]string{
		"1":              "per=1",
		"2":              "per=2",
		"3":              "per=3",
		"A":              "per=A",
		"BEINONI":        "tense=BEINONI",
		"D":              "num=D",
		"DP":             "num=D|num=P",
		"F":              "gen=F",
		"FUTURE":         "tense=FUTURE",
		"IMPERATIVE":     "tense=IMPERATIVE",
		"M":              "gen=M",
		"MF":             "gen=M|gen=F",
		"SP":             "num=s|num=P",
		"NEGATIVE":       "polar=neg",
		"P":              "num=P",
		"PAST":           "tense=PAST",
		"POSITIVE":       "polar=pos",
		"S":              "num=S",
		"PERS":           "type=PERS",
		"DEM":            "type=DEM",
		"REF":            "type=REF",
		"IMP":            "type=IMP",
		"INT":            "type=INT",
		"HIFIL":          "binyan=HIFIL",
		"PAAL":           "binyan=PAAL",
		"NIFAL":          "binyan=NIFAL",
		"HITPAEL":        "binyan=HITPAEL",
		"PIEL":           "binyan=PIEL",
		"PUAL":           "binyan=PUAL",
		"HUFAL":          "binyan=HUFAL",
		"TOINFINITIVE":   "type=TOINFINITIVE",
		"BAREINFINITIVE": "type=BAREINFINITIVE",
		"COORD":          "type=COORD",
		"SUB":            "type=SUB",
		"REL":            "type=REL",
		"SUBCONJ":        "type=SUBCONJ",
	}
	PP_FROM_MSR      map[string][]string
	PP_FROM_MSR_DATA = []string{
		// Based on Tsarfaty 2010 Relational-Realizational Parsing, p. 86
		"gen=F|gen=M|num=P|per=1|type=PERS:אנחנו",
		"gen=F|gen=M|num=S|per=1|type=PERS:אני",
		"gen=F|num=S|per=2|type=PERS:את",
		"gen=M|num=S|per=2|type=PERS:אתה",
		"gen=M|num=P|per=2|type=PERS:אתם",
		"gen=F|num=P|per=2|type=PERS:אתן",
		"gen=M|num=S|per=3|type=PERS:הוא",
		"gen=F|num=S|per=3|type=PERS:היא",
		"gen=M|num=P|per=3|type=PERS:הם",
		"gen=F|num=P|per=3|type=PERS:הן",
	}
	PP_BRIDGE = map[string]string{
		"CD": "של",
		"NN": "של",
		"VB": "את",
		"BN": "את",
		"IN": "",
	}
)

func init() {
	PP_FROM_MSR = make(map[string][]string, len(PP_FROM_MSR_DATA))
	for _, mapping := range PP_FROM_MSR_DATA {
		splitMap := strings.Split(mapping, ":")
		splitFeats := strings.Split(splitMap[0], FEATURE_PAIR_SEPARATOR)
		valuesStr := strings.Join(FeatureValues(splitFeats, true), FEATURE_SEPARATOR)
		valuesNoTypeStr := strings.Join(FeatureValues(splitFeats, false), FEATURE_SEPARATOR)
		if val, exists := PP_FROM_MSR[valuesStr]; exists {
			val = append(val, splitMap[1])
			PP_FROM_MSR[valuesStr] = val
		} else {
			PP_FROM_MSR[valuesStr] = []string{splitMap[1]}
		}
		if val, exists := PP_FROM_MSR[valuesNoTypeStr]; exists {
			val = append(val, splitMap[1])
			PP_FROM_MSR[valuesNoTypeStr] = val
		} else {
			PP_FROM_MSR[valuesNoTypeStr] = []string{splitMap[1]}
		}
	}
}

func FeatureValues(pairs []string, withType bool) []string {
	retval := make([]string, 0, len(pairs))
	var split []string
	for _, val := range pairs {
		split = strings.Split(val, FEATURE_VALUE_SEPARATOR)
		if withType || split[0] != "type" {
			retval = append(retval, split[1])
		}
	}
	return retval
}

type AnalyzedToken struct {
	Token     string
	Morphemes []types.BasicMorphemes
}

func (a *AnalyzedToken) NumMorphemes() (num int) {
	for _, m := range a.Morphemes {
		num += len(m)
	}
	return
}

func ParseMSR(msr string, add_suf bool) (string, string, map[string]string, string, error) {
	hostMSR := strings.Split(msr, FEATURE_SEPARATOR)
	sort.Strings(hostMSR[1:])
	featureMap := make(map[string]string, len(hostMSR)-1)
	resultMSR := make([]string, 0, len(hostMSR)-1)
	for _, msrFeatValue := range hostMSR[1:] {
		if lkpStr, exists := MSR_TYPE_FROM_VALUE[msrFeatValue]; exists {
			split := strings.Split(lkpStr, "=")
			if SKIP_BINYAN && len(split) > 0 && split[0] == "binyan" {
				continue
			}
			if add_suf {
				featureSplit := strings.Split(msrFeatValue, FEATURE_PAIR_SEPARATOR)
				for j, val := range featureSplit {
					featureSplit[j] = "suf_" + val
				}
				lkpStr = strings.Join(featureSplit, FEATURE_PAIR_SEPARATOR)
			}
			resultMSR = append(resultMSR, lkpStr)
			if len(split) == 2 {
				featureMap[split[0]] = split[1]
			} else {
				featureMap[split[0]] = msrFeatValue
			}
		} else {
			log.Println("Encountered unknown morph feature value", msrFeatValue, "- skipping")
		}
	}
	sort.Strings(resultMSR)
	return hostMSR[0], hostMSR[0], featureMap, strings.Join(resultMSR, FEATURE_PAIR_SEPARATOR), nil
}

func ParseMSRSuffix(hostPOS, msr string) (string, string, map[string]string, string, error) {
	hostMSR := strings.Split(msr, FEATURE_SEPARATOR)
	hostMSR = append(hostMSR, "PERS")
	feats := strings.Join(hostMSR[1:], FEATURE_SEPARATOR)
	var resultMorph string
	if suffixes, exists := PP_FROM_MSR[feats]; exists {
		resultMorph = suffixes[0]
	} else {
		resultMorph = "הם"
	}
	sort.Strings(hostMSR[1:])
	featureMap := make(map[string]string, len(hostMSR)-1)
	resultMSR := make([]string, 0, len(hostMSR)-1)
	for _, msrFeatValue := range hostMSR[1:] {
		if lkpStr, exists := MSR_TYPE_FROM_VALUE[msrFeatValue]; exists {
			split := strings.Split(lkpStr, "=")
			if SKIP_BINYAN && len(split) > 0 && split[0] == "binyan" {
				continue
			}
			resultMSR = append(resultMSR, lkpStr)
			if len(split) == 2 {
				featureMap[split[0]] = split[1]
			} else {
				featureMap[split[0]] = msrFeatValue
			}
		} else {
			log.Println("Encountered unknown morph feature value", msrFeatValue, "- skipping")
		}
	}
	sort.Strings(resultMSR)
	resultMSRStr := strings.Join(resultMSR, FEATURE_PAIR_SEPARATOR)
	var bridge string = ""
	if bridgeVal, exists := PP_BRIDGE[hostPOS]; exists {
		bridge = bridgeVal
	} else {
		log.Println("Encountered unknown POS for bridge", hostPOS)
	}
	return bridge, resultMorph, featureMap, resultMSRStr, nil
}

func ProcessAnalyzedToken(analysis string) (*AnalyzedToken, error) {
	var (
		split, msrs    []string
		curToken       *AnalyzedToken
		i              int
		curNode, curID int
		lemma          string
		def            bool
	)
	split = strings.Split(analysis, SEPARATOR)
	splitLen := len(split)
	if splitLen < 3 || splitLen%2 != 1 {
		return nil, errors.New("Wrong number of fields (" + analysis + ")")
	}
	curToken = &AnalyzedToken{
		Token:     split[0],
		Morphemes: make([]types.BasicMorphemes, 0, (splitLen-1)/2),
	}
	prefix := log.Prefix()
	log.SetPrefix(prefix + "Token " + curToken.Token + " ")
	for i = 1; i < splitLen; i += 2 {
		curNode, curID = 0, 0
		morphs := make(types.BasicMorphemes, 0, 4)
		msrs = strings.Split(split[i], MSR_SEPARATOR)
		lemma = split[i+1]
		def = false
		// Prefix morpheme (if exists)
		if len(msrs[0]) > 0 {
			if msrs[0] == "DEF" {
				def = true
			} else {
				return nil, errors.New("Unknown prefix MSR(" + msrs[0] + ")")
			}
		}
		if len(msrs[1]) == 0 {
			return nil, errors.New("Empty host MSR (" + analysis + ")")
		}
		// Host morpheme
		CPOS, POS, Features, FeatureStr, err := ParseMSR(msrs[1], false)
		if err != nil {
			return nil, err
		}
		if def {
			Features["def"] = "D"
		}
		hostMorph := &types.Morpheme{
			BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
			Form:              split[0],
			Lemma:             lemma,
			CPOS:              CPOS,
			POS:               POS,
			Features:          Features,
			TokenID:           0,
			FeatureStr:        FeatureStr,
		}
		morphs = append(morphs, hostMorph)
		curID++
		curNode++
		// Suffix morphemes
		if len(msrs[2]) > 0 {
			if CPOS == "NN" {
				// add prepositional pronoun features
				_, _, sufFeatures, sufFeatureStr, _ := ParseMSR(msrs[2], true)
				hostMorph.FeatureStr = strings.Join([]string{hostMorph.FeatureStr, sufFeatureStr}, FEATURE_PAIR_SEPARATOR)
				for k, v := range sufFeatures {
					hostMorph.Features[k] = v
				}
			} else if msrs[2][0] == '-' || (msrs[2][0] == 'S' && msrs[2][:5] != "S_ANP") {
				// add prepositional pronoun morphemes
				bridge, sufForm, sufFeatures, sufFeatureStr, err := ParseMSRSuffix(hostMorph.CPOS, msrs[2])
				if err != nil {
					return nil, err
				}
				if len(bridge) > 0 {
					morphs = append(morphs, &types.Morpheme{
						BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
						Form:              bridge,
						Lemma:             bridge,
						CPOS:              "POS",
						POS:               "POS",
						Features:          nil,
						TokenID:           0,
						FeatureStr:        "",
					})
					curID++
					curNode++
				}
				morphs = append(morphs, &types.Morpheme{
					BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
					Form:              sufForm,
					Lemma:             sufForm,
					CPOS:              "S_PRN",
					POS:               "S_PRN",
					Features:          sufFeatures,
					TokenID:           0,
					FeatureStr:        sufFeatureStr,
				})
				curID++
				curNode++
			}
		}
		curToken.Morphemes = append(curToken.Morphemes, morphs)
	}
	log.SetPrefix(prefix)
	return curToken, nil
}

func ProcessAnalyzedPrefix(analysis string) (*AnalyzedToken, error) {
	var (
		split, forms, prefix_msrs, msrs []string
		curToken                        *AnalyzedToken
		i                               int
		curNode, curID                  int
	)
	split = strings.Split(analysis, SEPARATOR)
	splitLen := len(split)
	if splitLen < 3 || splitLen%2 != 1 {
		return nil, errors.New("Wrong number of fields (" + analysis + ")")
	}
	curToken = &AnalyzedToken{
		Token:     split[0],
		Morphemes: make([]types.BasicMorphemes, 0, (splitLen-1)/2),
	}
	prefix := log.Prefix()
	log.SetPrefix(prefix + " Token " + curToken.Token)
	for i = 1; i < splitLen; i += 2 {
		curNode, curID = 0, 0
		morphs := make(types.BasicMorphemes, 0, 4)
		forms = strings.Split(split[i], PREFIX_SEPARATOR)
		prefix_msrs = strings.Split(split[i+1], PREFIX_MSR_SEPARATOR)
		if len(forms) != len(prefix_msrs) {
			return nil, errors.New("Mismatch between # of forms and # of MSRs (" + analysis + ")")
		}
		for j := 0; j < len(forms); j++ {
			msrs = strings.Split(prefix_msrs[j], MSR_SEPARATOR)
			// Add prefix morpheme
			if len(msrs[0]) > 0 {
				// replace -SUBCONJ for TEMP-SUBCONJ/REL-SUBCONJ
				morphs = append(morphs, &types.Morpheme{
					BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
					Form:              forms[j],
					Lemma:             forms[j],
					CPOS:              strings.Replace(msrs[0], "-SUBCONJ", "", -1),
					POS:               strings.Replace(msrs[0], "-SUBCONJ", "", -1),
					Features:          nil,
					TokenID:           0,
					FeatureStr:        "",
				})
				curID++
				curNode++
			}
		}
		curToken.Morphemes = append(curToken.Morphemes, morphs)
	}
	log.SetPrefix(prefix)
	return curToken, nil
}

type LexReader func(string) (*AnalyzedToken, error)

func Read(input io.Reader, format string) ([]*AnalyzedToken, error) {
	tokens := make([]*AnalyzedToken, 0, APPROX_LEX_SIZE)
	scan := bufio.NewScanner(input)
	var reader LexReader
	switch format {
	case "lexicon":
		reader = ProcessAnalyzedToken
	case "prefix":
		reader = ProcessAnalyzedPrefix
	default:
	}
	for scan.Scan() {
		line := scan.Text()
		token, err := reader(line)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
func ReadFile(filename string, format string) ([]*AnalyzedToken, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file, format)
}
