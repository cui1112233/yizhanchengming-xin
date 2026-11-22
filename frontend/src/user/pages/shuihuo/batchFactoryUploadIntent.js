export function batchFactoryUploadIntent(books, isUploaded) {
  const values = Array.isArray(books) ? books : [];
  const uploaded = values.filter(book => isUploaded(book));
  if (!values.length) return { kind: 'empty', uploaded, fresh: [] };
  if (!uploaded.length) return { kind: 'upload', uploaded, fresh: values };
  if (uploaded.length === values.length) return { kind: 'reupload', uploaded, fresh: [] };
  return {
    kind: 'mixed',
    uploaded,
    fresh: values.filter(book => !uploaded.includes(book))
  };
}
