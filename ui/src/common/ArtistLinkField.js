import React from 'react'
import PropTypes from 'prop-types'
import { Link } from 'react-admin'
import { withWidth } from '@material-ui/core'
import { useAlbumsPerPage } from './index'
import config from '../config'

export const useGetHandleArtistClick = (width) => {
  const [perPage] = useAlbumsPerPage(width)
  return (id) => {
    return config.devShowArtistPage && id !== config.variousArtistsId
      ? `/artist/${id}/show`
      : `/album?filter={"artist_id":"${id}"}&order=ASC&sort=max_year&displayedFilters={"compilation":true}&perPage=${perPage}`
  }
}

export const ArtistLinkField = withWidth()(({
  record,
  className,
  width,
  source,
}) => {
  const artistLink = useGetHandleArtistClick(width)

  const id = record[source + 'Id']
  let artistIds = []
  let artistNames = []
  if (id) {
    artistIds = id.split('/')
    artistNames = record[source].split('/')
  }
  return (
    <>
      {id &&
        artistIds.length === artistNames.length &&
        artistIds.slice(0, 5).map((artistId, index, arr) => (
          <>
            {index < 4 || index === artistIds.length - 1 ? (
              <Link
                to={artistLink(artistId)}
                onClick={(e) => e.stopPropagation()}
                className={className}
              >
                {artistNames[index]}
              </Link>
            ) : (
              <Link
                to={artistLink(artistId)}
                onClick={(e) => e.stopPropagation()}
                className={className}
              >
                ...
              </Link>
            )}
            {index < arr.length - 1 && <> / </>}
          </>
        ))}
      {(!id || artistIds.length !== artistNames.length) && record[source]}
    </>
  )
})

ArtistLinkField.propTypes = {
  record: PropTypes.object,
  className: PropTypes.string,
  source: PropTypes.string,
}

ArtistLinkField.defaultProps = {
  addLabel: true,
  source: 'albumArtist',
}
