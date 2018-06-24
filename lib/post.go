package fourfuse

import (
	"html"
	"os"
	"regexp"
	"strconv"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fourc "github.com/dedeibel/go-4chan-api/api"
	"golang.org/x/net/context"
)

const commentSlug string = "comment.txt"
const subjectSlug string = "subject.txt"

type Post struct {
	slug   string
	inode  uint64
	id     int64
	fourc  *fourc.Post
	image  *RemoteFile
	thread *Thread
}

func NewPost(fourc *fourc.Post, thread *Thread) *Post {
	slug := GeneratePostSlug(fourc)

	p := &Post{
		slug:   slug,
		inode:  hashs(HASH_POST_PREFIX + strconv.FormatInt(fourc.Id, 10)),
		id:     fourc.Id,
		fourc:  fourc,
		thread: thread}

	p.image = p.GetImage()

	return p
}

func GeneratePostSlug(post *fourc.Post) string {
	var slug = strconv.FormatInt(post.Id, 10)

	if !isEmptyString(post.Subject) {
		slug += " " + trimWhitespace(post.Subject)
	} else {
		cleanedComment := removeExifDataNotice(post.Comment)
		slug += " " + trimWhitespace(cleanedComment)
	}

	slug = html.UnescapeString(slug)
	slug = replaceHtmlMarkup(slug)
	slug = fsNameInvalidChars.ReplaceAllString(slug, "")
	slug = replaceMultipleSpaceByOne(slug)
	slug = strings.TrimSpace(slug)

	return truncateString(slug, fsNameMaxlen)
}

func removeExifDataNotice(comment string) string {
	return regexp.MustCompile("<span class=\"abbr\">\\[EXIF.*?(</table>)").ReplaceAllString(comment, "")
}

func (p *Post) Slug() string {
	return p.slug
}

func (p *Post) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = p.inode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (p *Post) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var dirDirs = make([]fuse.Dirent, 0, 3)

	dirDirs = append(dirDirs, p.getSubjectDirent())
	dirDirs = append(dirDirs, p.getCommentDirent())

	if p.HasImage() {
		dirDirs = append(dirDirs, p.image.GetDirent())
	}

	return dirDirs, nil
}

func (p *Post) getSubjectDirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: hashs(HASH_POST_INFO_PREFIX + p.slug + "sub"),
		Name:  subjectSlug,
		Type:  fuse.DT_File}
}

func (p *Post) getCommentDirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: hashs(HASH_POST_INFO_PREFIX + p.slug + "com"),
		Name:  commentSlug,
		Type:  fuse.DT_File}
}

func (p *Post) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == subjectSlug {
		return NewFile(
			hashs(HASH_POST_SUBJECT+p.slug),
			htmlToText(p.fourc.Subject)), nil
	} else if name == commentSlug {
		return NewFile(
			hashs(HASH_POST_COMMENT+p.slug),
			p.getCommentSanitized()), nil
	} else if p.HasImage() && name == p.image.Slug() {
		return p.image, nil
	} else {
		return nil, fuse.ENOENT
	}
}

func (p *Post) HasImage() bool {
	return p.image != nil
}

func (p *Post) getCommentSanitized() string {
	return htmlToText(p.fourc.Comment)
}

func (p *Post) GetImage() *RemoteFile {
	if p.fourc.File == nil {
		return nil
	}

	slug := p.postIdPathSegment() + " " + sanitizedFileName(p.fourc.File)

	file := NewRemoteFile(
		hashs(HASH_POST_IMAGE+p.slug),
		slug,
		p.fourc.ImageURL(),
		uint64(p.fourc.File.Size))

	return file
}

func (p *Post) GetThumbnail() *RemoteFile {
	if p.fourc.File == nil {
		return nil
	}

	slug := p.postIdPathSegment() + " " + sanitizedFileName(p.fourc.File)
	return p.buildThumbnailWithSlug(slug)
}

func (p *Post) postIdPathSegment() string {
	return sanitizePathSegment(strconv.FormatInt(p.id, 10))
}

func (p *Post) GetSamePrefixedSlugThumbnail() *RemoteFile {
	if p.fourc.File == nil {
		return nil
	}

	postLikeSlug := p.Slug() + "_t" + sanitizePathSegment(p.fourc.File.Ext)
	return p.buildThumbnailWithSlug(postLikeSlug)
}

func (p *Post) buildThumbnailWithSlug(slug string) *RemoteFile {
	return NewRemoteFile(
		hashs(HASH_POST_THUMBNAIL+p.slug),
		slug,
		p.fourc.ThumbURL(),
		0)
}

func (p *Post) GetDirent() fuse.Dirent {
	return fuse.Dirent{Inode: p.inode, Name: p.slug, Type: fuse.DT_Dir}
}
