// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables
import 'package:flutter/material.dart';
import 'theme.dart';
import 'data.dart';

// ═══════════════════════════════════════════════════════════════════════════
// ENTRY POINT
// ═══════════════════════════════════════════════════════════════════════════

void main() => runApp(const JobAggregatorApp());

class JobAggregatorApp extends StatefulWidget {
  const JobAggregatorApp({super.key});

  @override
  State<JobAggregatorApp> createState() => JobAggregatorAppState();

  static JobAggregatorAppState of(BuildContext context) =>
      context.findAncestorStateOfType<JobAggregatorAppState>()!;
}

class JobAggregatorAppState extends State<JobAggregatorApp> {
  ThemeMode _themeMode = ThemeMode.light;
  void toggleTheme() => setState(() {
        _themeMode = _themeMode == ThemeMode.light ? ThemeMode.dark : ThemeMode.light;
      });

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Sarkari Job Hub',
      debugShowCheckedModeBanner: false,
      theme: lightTheme,
      darkTheme: darkTheme,
      themeMode: _themeMode,
      home: const MainShell(),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// MAIN SHELL — Bottom Navigation
// ═══════════════════════════════════════════════════════════════════════════

class MainShell extends StatefulWidget {
  const MainShell({super.key});
  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _idx = 0;
  final _screens = const [HomeScreen(), MyURLsScreen(), NotificationsScreen(), ProfileScreen()];

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      body: IndexedStack(index: _idx, children: _screens),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: c.bgDefault,
          border: Border(top: BorderSide(color: c.borderSecondary, width: 1)),
        ),
        child: BottomNavigationBar(
          currentIndex: _idx,
          onTap: (i) => setState(() => _idx = i),
          type: BottomNavigationBarType.fixed,
          backgroundColor: c.bgDefault,
          selectedItemColor: AppColors.primary500,
          unselectedItemColor: c.textTertiary,
          selectedFontSize: 12,
          unselectedFontSize: 12,
          elevation: 0,
          items: const [
            BottomNavigationBarItem(icon: Icon(Icons.home_rounded), label: 'Home'),
            BottomNavigationBarItem(icon: Icon(Icons.link_rounded), label: 'My URLs'),
            BottomNavigationBarItem(icon: Icon(Icons.notifications_rounded), label: 'Alerts'),
            BottomNavigationBarItem(icon: Icon(Icons.person_rounded), label: 'Profile'),
          ],
        ),
      ),
      floatingActionButton: _idx == 1
          ? FloatingActionButton(
              onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => AddURLScreen())),
              backgroundColor: AppColors.primary500,
              child: Icon(Icons.add_rounded, color: Colors.white),
            )
          : null,
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// HOME SCREEN — Jobs from user's watched URLs
// ═══════════════════════════════════════════════════════════════════════════

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});
  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedChip = 0;

  List<String> get _sourceLabels {
    final labels = ['All'];
    for (final u in sampleWatchedURLs) {
      labels.add(u.label.split(' ').first); // "NIC", "SSC", etc.
    }
    return labels;
  }

  List<JobData> get _filteredJobs {
    if (_selectedChip == 0) return sampleJobs;
    final url = sampleWatchedURLs[_selectedChip - 1];
    return sampleJobs.where((j) => j.sourceId == url.id).toList();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final activeURLs = sampleWatchedURLs.where((u) => u.status == URLStatus.active).length;
    final totalJobs = sampleJobs.length;

    return SafeArea(
      child: CustomScrollView(
        slivers: [
          // Header
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text('Hello, User 👋', style: TextStyle(fontSize: 14, color: c.textSecondary)),
                    SizedBox(height: 4),
                    Text('Your Job Feed', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: c.textPrimary)),
                  ]),
                  Row(children: [
                    _ThemeToggle(),
                    SizedBox(width: 8),
                    Container(
                      decoration: BoxDecoration(
                        gradient: LinearGradient(colors: [AppColors.primary500, AppColors.primary400]),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: IconButton(icon: Icon(Icons.filter_list_rounded, color: Colors.white, size: 22), onPressed: () {}),
                    ),
                  ]),
                ],
              ),
            ),
          ),

          // Search
          SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.fromLTRB(20, 20, 20, 0),
              child: Container(
                decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(16), border: Border.all(color: c.borderPrimary)),
                child: TextField(
                  style: TextStyle(color: c.textPrimary, fontSize: 14),
                  decoration: InputDecoration(
                    hintText: 'Search your jobs...', hintStyle: TextStyle(color: c.textMuted, fontSize: 14),
                    prefixIcon: Icon(Icons.search_rounded, color: c.textMuted), border: InputBorder.none,
                    contentPadding: EdgeInsets.symmetric(vertical: 14, horizontal: 16),
                  ),
                ),
              ),
            ),
          ),

          // Stats
          SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.fromLTRB(20, 20, 20, 0),
              child: Row(children: [
                Expanded(child: _StatCard(icon: Icons.link_rounded, label: 'Watching', value: '$activeURLs URLs', gradient: [AppColors.primary500, AppColors.primary400])),
                SizedBox(width: 12),
                Expanded(child: _StatCard(icon: Icons.work_rounded, label: 'Total Jobs', value: '$totalJobs', gradient: [AppColors.info, AppColors.primary300])),
              ]),
            ),
          ),

          // Source Chips
          SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.only(top: 24),
              child: SizedBox(
                height: 44,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal, padding: EdgeInsets.symmetric(horizontal: 20),
                  itemCount: _sourceLabels.length,
                  itemBuilder: (ctx, i) {
                    final sel = _selectedChip == i;
                    return Padding(
                      padding: EdgeInsets.only(right: 8),
                      child: GestureDetector(
                        onTap: () => setState(() => _selectedChip = i),
                        child: AnimatedContainer(
                          duration: Duration(milliseconds: 200),
                          padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                          decoration: BoxDecoration(
                            color: sel ? AppColors.primary500 : c.bgSecondary,
                            borderRadius: BorderRadius.circular(12),
                            border: sel ? null : Border.all(color: c.borderSecondary),
                            boxShadow: sel ? [BoxShadow(color: AppColors.primary500.withValues(alpha: 0.3), blurRadius: 8, offset: Offset(0, 3))] : [],
                          ),
                          child: Text(_sourceLabels[i], style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: sel ? Colors.white : c.textPrimary)),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ),
          ),

          // Section Title
          SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.fromLTRB(20, 24, 20, 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Latest Opportunities', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: c.textPrimary)),
                  Text('${_filteredJobs.length} found', style: TextStyle(fontSize: 13, color: c.textSecondary)),
                ],
              ),
            ),
          ),

          // Job Cards
          SliverPadding(
            padding: EdgeInsets.symmetric(horizontal: 20),
            sliver: SliverList(
              delegate: SliverChildBuilderDelegate(
                (ctx, i) => _SlideIn(index: i, child: Padding(padding: EdgeInsets.only(bottom: 16), child: _JobCard(job: _filteredJobs[i]))),
                childCount: _filteredJobs.length,
              ),
            ),
          ),
          SliverToBoxAdapter(child: SizedBox(height: 16)),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// MY URLS SCREEN
// ═══════════════════════════════════════════════════════════════════════════

class MyURLsScreen extends StatelessWidget {
  const MyURLsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return SafeArea(
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Padding(
          padding: EdgeInsets.all(20),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('My Watched URLs', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: c.textPrimary)),
            SizedBox(height: 4),
            Text('${sampleWatchedURLs.length} sources being monitored', style: TextStyle(fontSize: 14, color: c.textSecondary)),
          ]),
        ),
        Expanded(
          child: ListView.builder(
            padding: EdgeInsets.symmetric(horizontal: 20),
            itemCount: sampleWatchedURLs.length,
            itemBuilder: (ctx, i) => _SlideIn(index: i, child: _URLCard(url: sampleWatchedURLs[i])),
          ),
        ),
      ]),
    );
  }
}

class _URLCard extends StatelessWidget {
  final WatchedURL url;
  const _URLCard({required this.url});

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final statusColor = switch (url.status) {
      URLStatus.active => AppColors.success,
      URLStatus.scraping => AppColors.info,
      URLStatus.failed => AppColors.error,
      URLStatus.paused => AppColors.warning,
    };
    final statusLabel = switch (url.status) {
      URLStatus.active => 'Active',
      URLStatus.scraping => 'Scraping...',
      URLStatus.failed => 'Failed (${url.failCount}x)',
      URLStatus.paused => 'Paused',
    };

    return Container(
      margin: EdgeInsets.only(bottom: 12),
      padding: EdgeInsets.all(16),
      decoration: c.cardDecoration,
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          // Emoji icon
          Container(
            width: 48, height: 48,
            decoration: BoxDecoration(gradient: LinearGradient(colors: [AppColors.primary500, AppColors.primary400]), borderRadius: BorderRadius.circular(12)),
            child: Center(child: Text(url.emoji, style: TextStyle(fontSize: 24))),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(url.label, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: c.textPrimary)),
              SizedBox(height: 2),
              Text(url.url, style: TextStyle(fontSize: 11, color: c.textMuted), maxLines: 1, overflow: TextOverflow.ellipsis),
            ]),
          ),
          // Status badge
          Container(
            padding: EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(color: statusColor.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
            child: Text(statusLabel, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: statusColor)),
          ),
        ]),
        SizedBox(height: 12),
        // Bottom info row
        Row(children: [
          _MiniChip(icon: Icons.work_rounded, label: '${url.jobCount} jobs', context: context),
          SizedBox(width: 8),
          _MiniChip(icon: Icons.schedule_rounded, label: url.lastScraped, context: context),
          Spacer(),
          // Actions
          Icon(url.status == URLStatus.paused ? Icons.play_arrow_rounded : Icons.pause_rounded, size: 20, color: c.textMuted),
          SizedBox(width: 12),
          Icon(Icons.delete_outline_rounded, size: 20, color: AppColors.error.withValues(alpha: 0.7)),
        ]),
      ]),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// ADD URL SCREEN
// ═══════════════════════════════════════════════════════════════════════════

class AddURLScreen extends StatefulWidget {
  const AddURLScreen({super.key});
  @override
  State<AddURLScreen> createState() => _AddURLScreenState();
}

class _AddURLScreenState extends State<AddURLScreen> {
  final _urlController = TextEditingController();
  final _labelController = TextEditingController();
  bool _isValid = false;

  void _validate() {
    setState(() {
      _isValid = _urlController.text.contains('.') && _labelController.text.trim().isNotEmpty;
    });
  }

  @override
  void dispose() { _urlController.dispose(); _labelController.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      appBar: AppBar(
        title: Text('Add URL to Watch', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: c.bgDefault,
        surfaceTintColor: Colors.transparent,
      ),
      body: SingleChildScrollView(
        padding: EdgeInsets.all(20),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          // Info banner
          Container(
            padding: EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.info.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.info.withValues(alpha: 0.3)),
            ),
            child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Icon(Icons.info_outline_rounded, color: AppColors.info, size: 20),
              SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Add any government job portal URL. Our scraper will check it every 6 hours and notify you of new listings.',
                  style: TextStyle(fontSize: 13, color: AppColors.info, height: 1.5),
                ),
              ),
            ]),
          ),

          SizedBox(height: 28),

          // URL field
          Text('Website URL', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: c.textPrimary)),
          SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderPrimary)),
            child: TextField(
              controller: _urlController,
              onChanged: (_) => _validate(),
              style: TextStyle(color: c.textPrimary, fontSize: 14),
              decoration: InputDecoration(
                hintText: 'https://recruitment.nic.in/...', hintStyle: TextStyle(color: c.textMuted),
                prefixIcon: Icon(Icons.link_rounded, color: AppColors.primary500),
                border: InputBorder.none, contentPadding: EdgeInsets.symmetric(vertical: 14, horizontal: 16),
              ),
              keyboardType: TextInputType.url,
            ),
          ),

          SizedBox(height: 20),

          // Label field
          Text('Friendly Name', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: c.textPrimary)),
          SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderPrimary)),
            child: TextField(
              controller: _labelController,
              onChanged: (_) => _validate(),
              style: TextStyle(color: c.textPrimary, fontSize: 14),
              decoration: InputDecoration(
                hintText: 'e.g., SSC Official, NIC Portal', hintStyle: TextStyle(color: c.textMuted),
                prefixIcon: Icon(Icons.label_outline_rounded, color: AppColors.primary500),
                border: InputBorder.none, contentPadding: EdgeInsets.symmetric(vertical: 14, horizontal: 16),
              ),
            ),
          ),

          SizedBox(height: 28),

          // Popular URLs section
          Text('Popular URLs', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: c.textPrimary)),
          SizedBox(height: 12),
          ..._popularURLs.map((p) => _PopularURLTile(
                label: p['label']!,
                url: p['url']!,
                emoji: p['emoji']!,
                onTap: () {
                  _urlController.text = p['url']!;
                  _labelController.text = p['label']!;
                  _validate();
                },
              )),

          SizedBox(height: 32),

          // Submit button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _isValid ? () => Navigator.pop(context) : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary500,
                foregroundColor: Colors.white,
                disabledBackgroundColor: c.bgTertiary,
                disabledForegroundColor: c.textMuted,
                padding: EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                elevation: 0,
              ),
              child: Text('Start Watching', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ),
          ),

          SizedBox(height: 12),
          Center(
            child: Text('Next scrape cycle runs in ~2 hours', style: TextStyle(fontSize: 12, color: c.textMuted)),
          ),
        ]),
      ),
    );
  }
}

final _popularURLs = [
  {'label': 'NIC Recruitment', 'url': 'https://recruitment.nic.in/index_new.php', 'emoji': '🏛️'},
  {'label': 'SSC Official', 'url': 'https://ssc.nic.in', 'emoji': '📋'},
  {'label': 'Railway RRB', 'url': 'https://rrbcdg.gov.in', 'emoji': '🚂'},
  {'label': 'UPSC', 'url': 'https://upsc.gov.in', 'emoji': '⚖️'},
  {'label': 'Indian Army', 'url': 'https://joinindianarmy.nic.in', 'emoji': '⭐'},
];

class _PopularURLTile extends StatelessWidget {
  final String label, url, emoji;
  final VoidCallback onTap;
  const _PopularURLTile({required this.label, required this.url, required this.emoji, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: EdgeInsets.only(bottom: 8),
        padding: EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderSecondary)),
        child: Row(children: [
          Text(emoji, style: TextStyle(fontSize: 20)),
          SizedBox(width: 12),
          Expanded(
            child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(label, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: c.textPrimary)),
              Text(url, style: TextStyle(fontSize: 11, color: c.textMuted), maxLines: 1, overflow: TextOverflow.ellipsis),
            ]),
          ),
          Icon(Icons.add_circle_outline_rounded, color: AppColors.primary500, size: 22),
        ]),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// JOB DETAIL SCREEN
// ═══════════════════════════════════════════════════════════════════════════

class JobDetailScreen extends StatelessWidget {
  final JobData job;
  const JobDetailScreen({super.key, required this.job});

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      body: CustomScrollView(slivers: [
        SliverAppBar(
          expandedHeight: 200, pinned: true, backgroundColor: AppColors.primary600,
          leading: IconButton(icon: Icon(Icons.arrow_back_rounded, color: Colors.white), onPressed: () => Navigator.pop(context)),
          actions: [
            IconButton(icon: Icon(Icons.share_rounded, color: Colors.white), onPressed: () {}),
            IconButton(icon: Icon(Icons.bookmark_border_rounded, color: Colors.white), onPressed: () {}),
          ],
          flexibleSpace: FlexibleSpaceBar(
            background: Container(
              decoration: BoxDecoration(gradient: LinearGradient(colors: [AppColors.primary500, AppColors.primary700], begin: Alignment.topCenter, end: Alignment.bottomCenter)),
              child: Center(child: Text(job.logo, style: TextStyle(fontSize: 80))),
            ),
          ),
        ),
        SliverToBoxAdapter(
          child: Padding(padding: EdgeInsets.all(20), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(job.title, style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: c.textPrimary)),
            SizedBox(height: 6),
            Text(job.organization, style: TextStyle(fontSize: 16, color: c.textSecondary)),
            SizedBox(height: 20),
            Row(children: [
              Expanded(child: _InfoBox(icon: Icons.people_rounded, label: 'Posts', value: job.posts, c: c)),
              SizedBox(width: 10),
              Expanded(child: _InfoBox(icon: Icons.account_balance_wallet_rounded, label: 'Salary', value: job.salary, c: c)),
              SizedBox(width: 10),
              Expanded(child: _InfoBox(icon: Icons.schedule_rounded, label: 'Deadline', value: job.deadline, c: c)),
            ]),
            SizedBox(height: 28),
            _Section(title: 'Key Highlights', items: job.highlights, c: c),
            SizedBox(height: 24),
            _Section(title: 'Eligibility', items: job.eligibility, c: c),
            SizedBox(height: 24),
            _Section(title: 'Important Dates', items: job.importantDates, c: c),
            SizedBox(height: 24),
            _Section(title: 'Selection Process', items: job.selectionProcess, c: c),
            SizedBox(height: 100),
          ])),
        ),
      ]),
      bottomNavigationBar: Container(
        padding: EdgeInsets.fromLTRB(20, 12, 20, 24),
        decoration: BoxDecoration(color: c.bgDefault, border: Border(top: BorderSide(color: c.borderSecondary))),
        child: Row(children: [
          Expanded(
            child: ElevatedButton(
              onPressed: () {},
              style: ElevatedButton.styleFrom(backgroundColor: AppColors.primary500, foregroundColor: Colors.white, padding: EdgeInsets.symmetric(vertical: 16), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)), elevation: 0),
              child: Text('Apply Now', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            ),
          ),
          SizedBox(width: 12),
          Container(
            decoration: BoxDecoration(border: Border.all(color: AppColors.primary500), borderRadius: BorderRadius.circular(12)),
            child: IconButton(icon: Icon(Icons.download_rounded, color: AppColors.primary500), onPressed: () {}),
          ),
        ]),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// NOTIFICATIONS SCREEN
// ═══════════════════════════════════════════════════════════════════════════

class NotificationsScreen extends StatelessWidget {
  const NotificationsScreen({super.key});
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final icons = [Icons.new_releases_rounded, Icons.warning_rounded, Icons.schedule_rounded, Icons.check_circle_rounded];
    final colors = [AppColors.info, AppColors.error, AppColors.warning, AppColors.success];

    return SafeArea(
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Padding(padding: EdgeInsets.all(20), child: Text('Notifications', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: c.textPrimary))),
        Expanded(
          child: ListView.builder(
            padding: EdgeInsets.symmetric(horizontal: 20),
            itemCount: sampleNotifications.length,
            itemBuilder: (ctx, i) {
              final n = sampleNotifications[i];
              return _SlideIn(
                index: i,
                child: Container(
                  margin: EdgeInsets.only(bottom: 12), padding: EdgeInsets.all(16),
                  decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderSecondary)),
                  child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Container(
                      padding: EdgeInsets.all(12),
                      decoration: BoxDecoration(color: colors[i % colors.length].withValues(alpha: 0.1), borderRadius: BorderRadius.circular(12)),
                      child: Icon(icons[i % icons.length], color: colors[i % colors.length], size: 22),
                    ),
                    SizedBox(width: 14),
                    Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text(n.title, style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: c.textPrimary)),
                      SizedBox(height: 4),
                      Text(n.description, style: TextStyle(fontSize: 12, color: c.textSecondary)),
                      SizedBox(height: 6),
                      Text(n.timestamp, style: TextStyle(fontSize: 11, color: c.textMuted)),
                    ])),
                  ]),
                ),
              );
            },
          ),
        ),
      ]),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// PROFILE SCREEN
// ═══════════════════════════════════════════════════════════════════════════

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final menus = [
      (Icons.settings_rounded, 'Settings'),
      (Icons.help_outline_rounded, 'Help & Support'),
      (Icons.lock_outline_rounded, 'Privacy Policy'),
      (Icons.logout_rounded, 'Logout'),
    ];
    return SafeArea(
      child: SingleChildScrollView(
        padding: EdgeInsets.all(20),
        child: Column(children: [
          SizedBox(height: 20),
          Container(
            padding: EdgeInsets.all(4),
            decoration: BoxDecoration(shape: BoxShape.circle, gradient: LinearGradient(colors: [AppColors.primary500, AppColors.primary400])),
            child: Container(width: 96, height: 96, decoration: BoxDecoration(color: c.bgDefault, shape: BoxShape.circle), child: Icon(Icons.person_rounded, size: 48, color: AppColors.primary500)),
          ),
          SizedBox(height: 16),
          Text('John Doe', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: c.textPrimary)),
          SizedBox(height: 4),
          Text('john.doe@email.com', style: TextStyle(fontSize: 14, color: c.textSecondary)),
          SizedBox(height: 8),
          // Device ID badge
          Container(
            padding: EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(color: c.bgTertiary, borderRadius: BorderRadius.circular(8)),
            child: Text('Device: a1b2c3d4', style: TextStyle(fontSize: 11, color: c.textMuted, fontFamily: 'monospace')),
          ),
          SizedBox(height: 32),
          ...menus.map((m) => Padding(
                padding: EdgeInsets.only(bottom: 12),
                child: Container(
                  padding: EdgeInsets.all(16),
                  decoration: BoxDecoration(color: c.bgDefault, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderSecondary)),
                  child: Row(children: [
                    Icon(m.$1, size: 22, color: AppColors.primary500),
                    SizedBox(width: 14),
                    Expanded(child: Text(m.$2, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500, color: c.textPrimary))),
                    Icon(Icons.arrow_forward_ios_rounded, size: 16, color: c.textMuted),
                  ]),
                ),
              )),
        ]),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// SHARED WIDGETS
// ═══════════════════════════════════════════════════════════════════════════

class _ThemeToggle extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Container(
      decoration: BoxDecoration(color: c.bgTertiary, borderRadius: BorderRadius.circular(12)),
      child: IconButton(
        icon: Icon(context.isDark ? Icons.light_mode_rounded : Icons.dark_mode_rounded, color: context.isDark ? AppColors.primary300 : AppColors.primary600, size: 22),
        onPressed: () => JobAggregatorApp.of(context).toggleTheme(),
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final IconData icon; final String label, value; final List<Color> gradient;
  const _StatCard({required this.icon, required this.label, required this.value, required this.gradient});
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(16),
      decoration: BoxDecoration(gradient: LinearGradient(colors: gradient, begin: Alignment.topLeft, end: Alignment.bottomRight), borderRadius: BorderRadius.circular(16)),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Icon(icon, color: Colors.white.withValues(alpha: 0.9), size: 28),
        SizedBox(height: 12),
        Text(value, style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white)),
        SizedBox(height: 4),
        Text(label, style: TextStyle(fontSize: 13, color: Colors.white.withValues(alpha: 0.85))),
      ]),
    );
  }
}

class _SlideIn extends StatelessWidget {
  final int index; final Widget child;
  const _SlideIn({required this.index, required this.child});
  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<double>(
      duration: Duration(milliseconds: 300 + (index * 100)),
      tween: Tween(begin: 0.0, end: 1.0), curve: Curves.easeOut,
      builder: (ctx, v, ch) => Transform.translate(offset: Offset(0, 20 * (1 - v)), child: Opacity(opacity: v, child: ch)),
      child: child,
    );
  }
}

class _JobCard extends StatelessWidget {
  final JobData job;
  const _JobCard({required this.job});
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final statusColor = switch (job.status) { JobStatus.isNew => AppColors.success, JobStatus.trending => AppColors.warning, JobStatus.hot => AppColors.error };
    final statusLabel = switch (job.status) { JobStatus.isNew => 'New', JobStatus.trending => 'Trending', JobStatus.hot => 'Hot' };
    // Find source URL label
    final sourceURL = sampleWatchedURLs.where((u) => u.id == job.sourceId).firstOrNull;

    return GestureDetector(
      onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => JobDetailScreen(job: job))),
      child: Container(
        padding: EdgeInsets.all(20), decoration: c.cardDecoration,
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            Container(
              width: 60, height: 60,
              decoration: BoxDecoration(gradient: LinearGradient(colors: job.gradient.cast<Color>()), borderRadius: BorderRadius.circular(16)),
              child: Center(child: Text(job.logo, style: TextStyle(fontSize: 28))),
            ),
            SizedBox(width: 14),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(job.title, style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: c.textPrimary)),
              SizedBox(height: 3),
              Text(job.organization, style: TextStyle(fontSize: 13, color: c.textSecondary)),
            ])),
            Container(
              padding: EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(color: statusColor.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
              child: Text(statusLabel, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: statusColor)),
            ),
          ]),
          SizedBox(height: 14),
          // Source badge
          if (sourceURL != null)
            Container(
              margin: EdgeInsets.only(bottom: 10),
              padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(color: AppColors.primary500.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
              child: Text('from ${sourceURL.label}', style: TextStyle(fontSize: 11, color: AppColors.primary500, fontWeight: FontWeight.w500)),
            ),
          Wrap(spacing: 8, runSpacing: 8, children: [
            _MiniChip(icon: Icons.people_rounded, label: job.posts, context: context),
            _MiniChip(icon: Icons.account_balance_wallet_rounded, label: job.salary, context: context),
          ]),
          SizedBox(height: 12),
          Container(
            padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(color: AppColors.warning.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
            child: Row(mainAxisSize: MainAxisSize.min, children: [
              Icon(Icons.schedule_rounded, size: 16, color: AppColors.warning),
              SizedBox(width: 6),
              Text('Deadline: ${job.deadline}', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: AppColors.warning)),
            ]),
          ),
        ]),
      ),
    );
  }
}

class _MiniChip extends StatelessWidget {
  final IconData icon; final String label; final BuildContext context;
  const _MiniChip({required this.icon, required this.label, required this.context});
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(color: c.bgTertiary, borderRadius: BorderRadius.circular(8)),
      child: Row(mainAxisSize: MainAxisSize.min, children: [
        Icon(icon, size: 15, color: c.textSecondary),
        SizedBox(width: 5),
        Text(label, style: TextStyle(fontSize: 12, color: c.textSecondary)),
      ]),
    );
  }
}

class _InfoBox extends StatelessWidget {
  final IconData icon; final String label, value; final AppThemeColors c;
  const _InfoBox({required this.icon, required this.label, required this.value, required this.c});
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(12),
      decoration: BoxDecoration(color: c.bgSecondary, borderRadius: BorderRadius.circular(12), border: Border.all(color: c.borderSecondary)),
      child: Column(children: [
        Icon(icon, size: 22, color: AppColors.primary500),
        SizedBox(height: 8),
        Text(value, style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: c.textPrimary), textAlign: TextAlign.center, maxLines: 2, overflow: TextOverflow.ellipsis),
        SizedBox(height: 4),
        Text(label, style: TextStyle(fontSize: 11, color: c.textSecondary)),
      ]),
    );
  }
}

class _Section extends StatelessWidget {
  final String title; final List<String> items; final AppThemeColors c;
  const _Section({required this.title, required this.items, required this.c});
  @override
  Widget build(BuildContext context) {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(title, style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: c.textPrimary)),
      SizedBox(height: 12),
      ...items.map((item) => Padding(
            padding: EdgeInsets.only(bottom: 10),
            child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Container(width: 6, height: 6, margin: EdgeInsets.only(top: 6, right: 12), decoration: BoxDecoration(color: AppColors.primary500, shape: BoxShape.circle)),
              Expanded(child: Text(item, style: TextStyle(fontSize: 14, color: c.textSecondary, height: 1.5))),
            ]),
          )),
    ]);
  }
}
