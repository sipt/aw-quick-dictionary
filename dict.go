package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreServices
#include "dict_service.h"
#include <stdlib.h>

// 直接包含实现代码
#import <Foundation/Foundation.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Forward declarations for Dictionary Services framework
NSArray *DCSCopyAvailableDictionaries();
NSArray *DCSCopyRecordsForSearchString(DCSDictionaryRef dictionary,
                                       CFStringRef string, void *, void *);
NSArray *DCSRecordCopyData(CFTypeRef record);
NSString *DCSDictionaryGetName(DCSDictionaryRef dictID);

DictResults *query_dictionary(const char *word, const char *dict_name) {
    @autoreleasepool {
        if (!word || !dict_name) {
            return NULL;
        }

        DictResults *results = malloc(sizeof(DictResults));
        if (!results) {
            return NULL;
        }

        results->results = NULL;
        results->count = 0;

        NSString *dictname = [NSString stringWithCString:dict_name encoding:NSUTF8StringEncoding];
        if (!dictname) {
            free(results);
            return NULL;
        }

        DCSDictionaryRef dictionary = NULL;
        NSArray *dicts = DCSCopyAvailableDictionaries();

        // Find the target dictionary
        for (NSObject *aDict in dicts) {
            NSString *aShortName = DCSDictionaryGetName((__bridge DCSDictionaryRef)aDict);
            if ([aShortName isEqualToString:dictname]) {
                dictionary = (__bridge DCSDictionaryRef)aDict;
                break;
            }
        }

        if (!dictionary) {
            free(results);
            return NULL;
        }

        // Convert word to NSString
        NSString *wordString = [NSString stringWithCString:word encoding:NSUTF8StringEncoding];
        if (!wordString) {
            free(results);
            return NULL;
        }

        // Query the dictionary
        NSArray *records = DCSCopyRecordsForSearchString(dictionary, (__bridge CFStringRef)wordString, 0, 0);

        if (!records || [records count] == 0) {
            free(results);
            return NULL;
        }

        // Allocate memory for results
        results->count = (int)[records count];
        results->results = malloc(sizeof(char*) * results->count);
        if (!results->results) {
            free(results);
            return NULL;
        }

        // Extract data from each record
        for (int i = 0; i < results->count; i++) {
            NSString *recordData = (NSString *)DCSRecordCopyData((__bridge CFTypeRef)[records objectAtIndex:i]);
            if (recordData) {
                const char *utf8String = [recordData UTF8String];
                size_t len = strlen(utf8String);
                results->results[i] = malloc(len + 1);
                if (results->results[i]) {
                    strcpy(results->results[i], utf8String);
                } else {
                    results->results[i] = NULL;
                }
            } else {
                results->results[i] = NULL;
            }
        }

        return results;
    }
}

DictResults *get_dictionary_names() {
    @autoreleasepool {
        DictResults *results = malloc(sizeof(DictResults));
        if (!results) {
            return NULL;
        }

        results->results = NULL;
        results->count = 0;

        // Get available dictionaries - returns NSSet, not NSArray
        NSSet *dictSet = (NSSet *)DCSCopyAvailableDictionaries();
        if (!dictSet || [dictSet count] == 0) {
            if (dictSet) [dictSet release];
            free(results);
            return NULL;
        }

        // Allocate memory for results
        results->count = (int)[dictSet count];
        results->results = malloc(sizeof(char*) * results->count);
        if (!results->results) {
            [dictSet release];
            free(results);
            return NULL;
        }

        // Extract dictionary names from set
        int i = 0;
        for (id dictRef in dictSet) {
            NSString *dictName = DCSDictionaryGetName((__bridge DCSDictionaryRef)dictRef);
            if (dictName) {
                const char *utf8String = [dictName UTF8String];
                if (utf8String) {
                    size_t len = strlen(utf8String);
                    results->results[i] = malloc(len + 1);
                    if (results->results[i]) {
                        strcpy(results->results[i], utf8String);
                    } else {
                        results->results[i] = NULL;
                    }
                } else {
                    results->results[i] = NULL;
                }
            } else {
                results->results[i] = NULL;
            }
            i++;
        }

        [dictSet release];
        return results;
    }
}

void free_dict_results(DictResults *results) {
    if (!results) {
        return;
    }

    if (results->results) {
        for (int i = 0; i < results->count; i++) {
            if (results->results[i]) {
                free(results->results[i]);
            }
        }
        free(results->results);
    }

    free(results);
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unsafe"

	aw "github.com/deanishe/awgo"
)

const WelcomeArt = `
 ▗▄▄▄▖ ▗▖ ▗▖▗▄▄▄▖ ▗▄▄▖▗▖ ▗▖   
 ▐▌ ▐▌ ▐▌ ▐▌  █  ▐▌   ▐▌▗▞▘   
 ▐▌ ▐▌ ▐▌ ▐▌  █  ▐▌   ▐▛▚▖    
 ▐▙▄▟▙▖▝▚▄▞▘▗▄█▄▖▝▚▄▄▖▐▌ ▐▌   
                                                                                 
 ▗▄▄▄ ▗▄▄▄▖ ▗▄▄▖▗▄▄▄▖▗▄▄▄▖ ▗▄▖ ▗▖  ▗▖ ▗▄▖ ▗▄▄▖▗▖  ▗▖
 ▐▌  █  █  ▐▌     █    █  ▐▌ ▐▌▐▛▚▖▐▌▐▌ ▐▌▐▌ ▐▌▝▚▞▘ 
 ▐▌  █  █  ▐▌     █    █  ▐▌ ▐▌▐▌ ▝▜▌▐▛▀▜▌▐▛▀▚▖ ▐▌  
 ▐▙▄▄▀▗▄█▄▖▝▚▄▄▖  █  ▗▄█▄▖▝▚▄▞▘▐▌  ▐▌▐▌ ▐▌▐▌ ▐▌ ▐▌  
`

var wf *aw.Workflow
var service *DictService

func init() {
	wf = aw.New()
	service = NewDictService()
}

// 默认词典名称
const DefaultDictName = "牛津英汉汉英词典"

// DictService 字典服务结构体
type DictService struct{}

// NewDictService 创建新的字典服务实例
func NewDictService() *DictService {
	return &DictService{}
}

// QueryWord 查询单词，返回结果数组
func (d *DictService) QueryWord(word string, dictName string) ([]string, error) {
	if word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	if dictName == "" {
		return nil, fmt.Errorf("dictionary name cannot be empty")
	}

	// 转换Go string到C string
	cWord := C.CString(word)
	defer C.free(unsafe.Pointer(cWord))

	cDictName := C.CString(dictName)
	defer C.free(unsafe.Pointer(cDictName))

	// 调用C函数
	cResults := C.query_dictionary(cWord, cDictName)
	if cResults == nil {
		return nil, fmt.Errorf("failed to query dictionary or no results found")
	}
	defer C.free_dict_results(cResults)

	// 转换C结果到Go slice
	count := int(cResults.count)
	if count == 0 {
		return nil, fmt.Errorf("no results found")
	}

	results := make([]string, count)

	// 获取C数组指针
	cArray := (*[1 << 30]*C.char)(unsafe.Pointer(cResults.results))[:count:count]

	for i := 0; i < count; i++ {
		if cArray[i] != nil {
			results[i] = C.GoString(cArray[i])
		}
	}

	return results, nil
}

// GetDictionaryNames 获取所有可用词典的名称列表
func (d *DictService) GetDictionaryNames() ([]string, error) {
	// 调用C函数
	cResults := C.get_dictionary_names()
	if cResults == nil {
		return nil, fmt.Errorf("failed to get dictionary names or no dictionaries found")
	}
	defer C.free_dict_results(cResults)

	// 转换C结果到Go slice
	count := int(cResults.count)
	if count == 0 {
		return nil, fmt.Errorf("no dictionaries found")
	}

	results := make([]string, count)

	// 获取C数组指针
	cArray := (*[1 << 30]*C.char)(unsafe.Pointer(cResults.results))[:count:count]

	for i := 0; i < count; i++ {
		if cArray[i] != nil {
			results[i] = C.GoString(cArray[i])
		}
	}

	return results, nil
}

func getDictNames(filter string) {
	wf.Run(func() {
		currentDictName := ""
		if currentDict, err := wf.Cache.Load("dictName"); err == nil && len(currentDict) > 0 {
			currentDictName = string(currentDict)
		}
		names, err := service.GetDictionaryNames()
		if err != nil {
			return
		}
		length := 0
		wf.NewItem("✓ " + currentDictName).Arg(currentDictName).Valid(false)
		sort.Strings(names)
		for _, name := range names {
			if filter == "" || strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
				title := name
				if currentDictName == name {
					continue
				}
				wf.NewItem(title).Arg(name).Valid(true)
				length += 1
			}
		}
		if length == 0 {
			wf.NewItem("No dictionaries found").Valid(false)
		}
		return
	})
	wf.SendFeedback()
}

func selectDictName(dictName string) {
	wf.Run(func() {
		err := wf.Cache.Store("dictName", []byte(dictName))
		if err != nil {
			log.Println("set [dictName] failed.", err.Error())
			wf.NewWarningItem("Set dictName failed", err.Error())
			return
		}
	})
}

type DictResult struct {
	Response string `json:"response"`
	Footer   string `json:"footer"`
}

func queryWord(word string) {
	wf.Run(func() {
		currentDictName := ""
		if currentDict, err := wf.Cache.Load("dictName"); err == nil && len(currentDict) > 0 {
			currentDictName = string(currentDict)
		}
		if currentDictName == "" {
			currentDictName = DefaultDictName
		}
		dictResult := DictResult{
			Response: "",
			Footer:   "From: " + currentDictName,
		}
		defer func() {
			bytes, _ := json.Marshal(dictResult)
			fmt.Printf("%s", string(bytes))
		}()
		if word == "" {
			// 提示：请输入单词后按下回车
			dictResult.Response = "Please input a word and press Enter." + "\n\n" + "```" + WelcomeArt + "```"
			return
		}
		results, err := service.QueryWord(word, currentDictName) // Assuming DictZh is a predefined constant for the default dictionary
		if err != nil {
			dictResult.Response = fmt.Sprintf("Error: %v\n", err)
			return
		}

		fallback := ""
		for _, result := range results {
			resultWord, mdContent, err := parseHtmlToMd(result)
			if err != nil {
				dictResult.Response = fmt.Sprintf("Error: %v\n", err)
				continue
			}
			if fallback == "" {
				fallback = mdContent
			}
			if resultWord == word {
				dictResult.Response = mdContent
				break
			}
		}
		if dictResult.Response == "" {
			dictResult.Response = fallback
		}
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("{\"response\": \"```%s```\"}", strings.ReplaceAll(WelcomeArt, "\n", "\\n"))
		return
	}

	op := os.Args[1]

	switch op {
	case "--list-dicts":
		filter := ""
		if len(os.Args) > 2 {
			filter = os.Args[2]
		}
		getDictNames(filter)
	case "--select-dict":
		selectDictName(os.Args[2])
	default:
		queryWord(op)
	}
}
