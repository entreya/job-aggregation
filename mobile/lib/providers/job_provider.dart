import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/database_manager.dart';
import '../models/job.dart';

// Notifier for the search query
class SearchQueryNotifier extends Notifier<String> {
  @override
  String build() {
    return '';
  }

  void set(String query) {
    state = query;
  }
}

final searchQueryProvider = NotifierProvider<SearchQueryNotifier, String>(SearchQueryNotifier.new);

// Provider for the list of jobs, dependent on search query
final jobsProvider = FutureProvider.autoDispose<List<Job>>((ref) async {
  final query = ref.watch(searchQueryProvider);
  final dbManager = DatabaseManager();
  
  // Search or get all
  final results = await dbManager.searchJobs(query);
  
  return results.map((map) => Job.fromMap(map)).toList();
});

// Provider to handle database updates (Pull-to-Refresh)
final databaseUpdateProvider = FutureProvider<bool>((ref) async {
  final dbManager = DatabaseManager();
  return await dbManager.checkForUpdates();
});
