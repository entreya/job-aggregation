// ignore_for_file: prefer_const_constructors
import 'theme.dart';

// ═══════════════════════════════════════════════════════════════════════════
// DATA MODELS
// ═══════════════════════════════════════════════════════════════════════════

enum URLStatus { active, scraping, failed, paused }
enum JobStatus { isNew, trending, hot }

class WatchedURL {
  final String id;
  final String url;
  final String label;
  final URLStatus status;
  final int jobCount;
  final String lastScraped;
  final int failCount;
  final String emoji;

  const WatchedURL({
    required this.id,
    required this.url,
    required this.label,
    required this.status,
    required this.jobCount,
    required this.lastScraped,
    this.failCount = 0,
    required this.emoji,
  });
}

class JobData {
  final String title;
  final String organization;
  final String posts;
  final String deadline;
  final String salary;
  final String logo;
  final List<dynamic> gradient;
  final JobStatus status;
  final String sourceId; // links to WatchedURL.id
  final List<String> highlights;
  final List<String> eligibility;
  final List<String> importantDates;
  final List<String> selectionProcess;

  const JobData({
    required this.title,
    required this.organization,
    required this.posts,
    required this.deadline,
    required this.salary,
    required this.logo,
    required this.gradient,
    required this.status,
    required this.sourceId,
    this.highlights = const [],
    this.eligibility = const [],
    this.importantDates = const [],
    this.selectionProcess = const [],
  });
}

class NotificationData {
  final String title;
  final String description;
  final String timestamp;
  final dynamic icon;
  final dynamic color;

  const NotificationData({
    required this.title,
    required this.description,
    required this.timestamp,
    required this.icon,
    required this.color,
  });
}

// ═══════════════════════════════════════════════════════════════════════════
// SAMPLE DATA — Watched URLs
// ═══════════════════════════════════════════════════════════════════════════

final List<WatchedURL> sampleWatchedURLs = [
  WatchedURL(
    id: 'url-nic',
    url: 'https://recruitment.nic.in/index_new.php',
    label: 'NIC Recruitment Portal',
    status: URLStatus.active,
    jobCount: 47,
    lastScraped: '2 hours ago',
    emoji: '🏛️',
  ),
  WatchedURL(
    id: 'url-ssc',
    url: 'https://ssc.nic.in',
    label: 'SSC Official',
    status: URLStatus.active,
    jobCount: 12,
    lastScraped: '2 hours ago',
    emoji: '📋',
  ),
  WatchedURL(
    id: 'url-railway',
    url: 'https://rrbcdg.gov.in',
    label: 'Railway Recruitment',
    status: URLStatus.active,
    jobCount: 8,
    lastScraped: '2 hours ago',
    emoji: '🚂',
  ),
  WatchedURL(
    id: 'url-upsc',
    url: 'https://upsc.gov.in',
    label: 'UPSC Official',
    status: URLStatus.failed,
    jobCount: 3,
    lastScraped: '6 hours ago',
    failCount: 3,
    emoji: '⚖️',
  ),
  WatchedURL(
    id: 'url-army',
    url: 'https://joinindianarmy.nic.in',
    label: 'Indian Army Agniveer',
    status: URLStatus.paused,
    jobCount: 5,
    lastScraped: '1 day ago',
    emoji: '⭐',
  ),
];

// ═══════════════════════════════════════════════════════════════════════════
// SAMPLE DATA — Jobs (linked to URLs via sourceId)
// ═══════════════════════════════════════════════════════════════════════════

final List<JobData> sampleJobs = [
  JobData(
    title: 'SSC CHSL 2025',
    organization: 'Staff Selection Commission',
    posts: '3,712 Posts',
    deadline: '15 Mar 2025',
    salary: '₹25,500 - ₹81,100',
    logo: '🏛️',
    gradient: [AppColors.primary500, AppColors.primary400],
    status: JobStatus.isNew,
    sourceId: 'url-ssc',
    highlights: ['📋 Online application mode', '🎓 12th Pass / Graduate', '📝 Computer Based Exam', '💰 Attractive salary with DA & HRA', '🏥 Medical benefits'],
    eligibility: ['🔹 Age: 18-27 years', '🔹 Education: 12th Pass', '🔹 Nationality: Indian', '🔹 Age relaxation: SC/ST 5 yrs, OBC 3 yrs'],
    importantDates: ['📅 Notification: 10 Jan 2025', '📅 Start: 15 Jan 2025', '📅 End: 15 Mar 2025', '📅 Exam: May 2025'],
    selectionProcess: ['1️⃣ CBE Tier-I', '2️⃣ CBE Tier-II', '3️⃣ Skill Test', '4️⃣ Document Verification'],
  ),
  JobData(
    title: 'RRB NTPC 2025',
    organization: 'Railway Recruitment Board',
    posts: '12,000+ Posts',
    deadline: '22 Mar 2025',
    salary: '₹35,400 - ₹1,12,400',
    logo: '🚂',
    gradient: [AppColors.info, AppColors.primary300],
    status: JobStatus.trending,
    sourceId: 'url-railway',
    highlights: ['📋 Online application', '🎓 Graduation required', '📝 Computer Based Exam', '💰 Railway benefits', '🏥 Full medical coverage'],
    eligibility: ['🔹 Age: 18-33 years', '🔹 Education: Graduation', '🔹 Nationality: Indian', '🔹 Physical fitness required'],
    importantDates: ['📅 Notification: 5 Jan 2025', '📅 Start: 10 Jan 2025', '📅 End: 22 Mar 2025', '📅 CBT-I: June 2025'],
    selectionProcess: ['1️⃣ CBT-I', '2️⃣ CBT-II', '3️⃣ Typing/CBAT', '4️⃣ Document Verification'],
  ),
  JobData(
    title: 'UPSC CSE 2025',
    organization: 'Union Public Service Commission',
    posts: '1,000 Posts',
    deadline: '28 Feb 2025',
    salary: '₹56,100 - ₹2,50,000',
    logo: '⚖️',
    gradient: [AppColors.success, AppColors.success],
    status: JobStatus.hot,
    sourceId: 'url-upsc',
    highlights: ['📋 Online via upsconline.nic.in', '🎓 Any graduate', '📝 Prelims + Mains + Interview', '💰 Highest government salary', '🏛️ Most prestigious exam'],
    eligibility: ['🔹 Age: 21-32 years', '🔹 Education: Any graduation', '🔹 Nationality: Indian', '🔹 Attempts: 6 general, unlimited SC/ST'],
    importantDates: ['📅 Notification: 15 Jan 2025', '📅 Start: 20 Jan 2025', '📅 End: 28 Feb 2025', '📅 Prelims: 25 May 2025'],
    selectionProcess: ['1️⃣ Preliminary Exam', '2️⃣ Main Exam', '3️⃣ Interview', '4️⃣ Final Merit List'],
  ),
  JobData(
    title: 'Scientist/Engineer SC',
    organization: 'NIC - National Informatics Centre',
    posts: '150 Posts',
    deadline: '20 Mar 2025',
    salary: '₹56,100 - ₹1,77,500',
    logo: '💻',
    gradient: [AppColors.primary600, AppColors.primary400],
    status: JobStatus.isNew,
    sourceId: 'url-nic',
    highlights: ['📋 Online application', '🎓 B.E./B.Tech in CS/IT', '📝 Written Test + Interview', '💰 Central Govt Pay Scale', '🖥️ Work in e-Governance'],
    eligibility: ['🔹 Age: 21-30 years', '🔹 Education: B.E./B.Tech CS/IT', '🔹 Nationality: Indian', '🔹 GATE score preferred'],
    importantDates: ['📅 Notification: 1 Feb 2025', '📅 Start: 5 Feb 2025', '📅 End: 20 Mar 2025', '📅 Exam: April 2025'],
    selectionProcess: ['1️⃣ Written Examination', '2️⃣ Technical Interview', '3️⃣ Document Verification', '4️⃣ Final Selection'],
  ),
  JobData(
    title: 'Army Agniveer 2025',
    organization: 'Indian Army',
    posts: '45,000 Posts',
    deadline: '18 Mar 2025',
    salary: '₹30,000 - ₹40,000',
    logo: '⭐',
    gradient: [AppColors.error, AppColors.error],
    status: JobStatus.isNew,
    sourceId: 'url-army',
    highlights: ['📋 Online via joinindianarmy.nic.in', '🎓 10th/12th pass', '📝 Written + Physical test', '💰 Seva Nidhi package', '🎖️ Serve the nation'],
    eligibility: ['🔹 Age: 17.5-23 years', '🔹 Education: 10th/12th pass', '🔹 Nationality: Indian', '🔹 Physical standards required'],
    importantDates: ['📅 Notification: 20 Jan 2025', '📅 Start: 1 Feb 2025', '📅 End: 18 Mar 2025', '📅 Rally: April-May 2025'],
    selectionProcess: ['1️⃣ Online CEE', '2️⃣ Physical Fitness', '3️⃣ Medical Exam', '4️⃣ Document Verification'],
  ),
  JobData(
    title: 'Data Entry Operator',
    organization: 'NIC - National Informatics Centre',
    posts: '200 Posts',
    deadline: '25 Mar 2025',
    salary: '₹25,500 - ₹81,100',
    logo: '⌨️',
    gradient: [AppColors.primary500, AppColors.primary300],
    status: JobStatus.trending,
    sourceId: 'url-nic',
    highlights: ['📋 Online application', '🎓 12th pass + typing skill', '📝 Skill test', '💰 Government pay scale', '🖥️ Central government posting'],
    eligibility: ['🔹 Age: 18-27 years', '🔹 Education: 12th pass', '🔹 Typing: 35 WPM English', '🔹 Computer knowledge required'],
    importantDates: ['📅 Notification: 10 Feb 2025', '📅 Start: 15 Feb 2025', '📅 End: 25 Mar 2025', '📅 Exam: May 2025'],
    selectionProcess: ['1️⃣ Written Test', '2️⃣ Typing Test', '3️⃣ Document Verification', '4️⃣ Final Selection'],
  ),
];

final List<NotificationData> sampleNotifications = [
  NotificationData(
    title: '🆕 3 new jobs from NIC Portal',
    description: 'Scientist/Engineer SC and 2 more positions added from recruitment.nic.in',
    timestamp: '2 hours ago',
    icon: null, // will use Icons.new_releases_rounded
    color: null, // will use AppColors.info
  ),
  NotificationData(
    title: '⚠️ UPSC scraping failed',
    description: 'upsc.gov.in returned error 403. Will retry in next cycle.',
    timestamp: '6 hours ago',
    icon: null,
    color: null,
  ),
  NotificationData(
    title: 'SSC CHSL deadline approaching!',
    description: 'Only 5 days left to apply. Don\'t miss out!',
    timestamp: '1 day ago',
    icon: null,
    color: null,
  ),
  NotificationData(
    title: '✅ Railway scraping successful',
    description: 'Found 8 active job listings from rrbcdg.gov.in',
    timestamp: '2 days ago',
    icon: null,
    color: null,
  ),
];
