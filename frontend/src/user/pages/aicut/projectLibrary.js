export function projectEditorUrl(id) {
  return `/ai-cut?project=${encodeURIComponent(id)}`;
}

export function libraryProjects(projects, query = '', sort = 'recent') {
  const needle = query.trim().toLocaleLowerCase();
  return projects.filter(project => String(project.title || '').toLocaleLowerCase().includes(needle)).sort((a, b) =>
    sort === 'name' ? a.title.localeCompare(b.title, 'zh-CN') :
      (Date.parse(b.updatedAt || b.createdAt) || 0) - (Date.parse(a.updatedAt || a.createdAt) || 0));
}
