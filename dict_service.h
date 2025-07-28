#ifndef DICT_SERVICE_H
#define DICT_SERVICE_H

#ifdef __cplusplus
extern "C"
{
#endif

  // Structure to hold dictionary query results
  typedef struct
  {
    char **results;
    int count;
  } DictResults;

  // Query dictionary for a word and return results
  DictResults *query_dictionary(const char *word, const char *dict_name);

  // Get list of available dictionary names
  DictResults *get_dictionary_names();

  // Free the memory allocated for results
  void free_dict_results(DictResults *results);

#ifdef __cplusplus
}
#endif

#endif // DICT_SERVICE_H