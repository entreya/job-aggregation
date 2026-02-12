import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/job_provider.dart';
import 'job_card.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final jobsAsyncValue = ref.watch(jobsProvider);
    final searchQuery = ref.watch(searchQueryProvider);

    return Scaffold(
      body: NestedScrollView(
        headerSliverBuilder: (context, innerBoxIsScrolled) => [
          SliverAppBar.large(
            title: const Text('Job Aggregator'),
            floating: true,
            pinned: true,
            bottom: PreferredSize(
              preferredSize: const Size.fromHeight(60),
                child: SearchBar(
                  hintText: 'Search jobs...',
                  leading: const Icon(Icons.search),
                  onChanged: (value) {
                    ref.read(searchQueryProvider.notifier).set(value);
                  },
                ),
            ),
          ),
        ],
        body: RefreshIndicator(
          onRefresh: () async {
            // Check for DB updates
             await ref.refresh(databaseUpdateProvider.future);
             // Invalidate jobs provider to re-fetch from DB
             ref.invalidate(jobsProvider);
          },
          child: jobsAsyncValue.when(
            data: (jobs) {
              if (jobs.isEmpty) {
                return Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.search_off, size: 64, color: Colors.grey),
                      const SizedBox(height: 16),
                      Text(
                        searchQuery.isEmpty 
                            ? 'No jobs found. Pull to refresh!' 
                            : 'No results for "$searchQuery"',
                        style: Theme.of(context).textTheme.bodyLarge,
                      ),
                    ],
                  ),
                );
              }
              return ListView.builder(
                padding: EdgeInsets.zero,
                itemCount: jobs.length,
                itemBuilder: (context, index) {
                  return JobCard(job: jobs[index]);
                },
              );
            },
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (err, stack) => Center(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Text('Error: $err'),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
